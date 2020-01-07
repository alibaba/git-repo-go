// Copyright © 2019 Alibaba Co. Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"crypto/sha256"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/git-repo-go/cap"
	"github.com/aliyun/git-repo-go/config"
	"github.com/aliyun/git-repo-go/format"
	"github.com/aliyun/git-repo-go/version"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/openpgp"
	"gopkg.in/yaml.v2"
)

const (
	// defaultUpgradeURL indicates where to download git-repo new package
	defaultUpgradeURL = "http://repo.code.alibaba-inc.com/download"
)

type upgradeCommand struct {
	cmd        *cobra.Command
	httpClient *http.Client

	O struct {
		URL          string
		Test         bool
		Version      string
		NoCertChecks bool
	}
}

type upgradeVersion struct {
	Production string `yaml:"production"`
	Test       string `yaml:"test"`
}

func (v *upgradeCommand) Command() *cobra.Command {
	if v.cmd != nil {
		return v.cmd
	}

	v.cmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Check and upgrade git-repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Execute(args)
		},
	}

	v.cmd.Flags().StringVarP(&v.O.URL,
		"url",
		"u",
		defaultUpgradeURL,
		"upgrade from this URL")
	v.cmd.Flags().StringVar(&v.O.Version,
		"version",
		"",
		"install specific version")
	v.cmd.Flags().BoolVarP(&v.O.Test,
		"test",
		"t",
		false,
		"upgrade to test version")
	v.cmd.Flags().BoolVar(&v.O.NoCertChecks,
		"no-cert-checks",
		false,
		"Disable verifying ssl certs (unsafe)")

	return v.cmd
}

func (v *upgradeCommand) HTTPClient() *http.Client {
	var (
		timeout = time.Duration(10)
	)

	if v.httpClient != nil {
		return v.httpClient
	}

	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   timeout * time.Second,
			KeepAlive: timeout * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: v.O.NoCertChecks || config.NoCertChecks(),
		},
		TLSHandshakeTimeout:   timeout * time.Second,
		ResponseHeaderTimeout: timeout * time.Second,
		ExpectContinueTimeout: timeout * time.Second,
		MaxIdleConns:          10,
		IdleConnTimeout:       timeout * time.Second,
		DisableCompression:    true,
	}

	v.httpClient = &http.Client{Transport: tr}
	return v.httpClient
}

func (v upgradeCommand) GetUpgradeVersion() (string, error) {
	var (
		targetVersion = upgradeVersion{}
	)

	vURL := v.O.URL + "/version.yml"

	req, err := http.NewRequest("GET", vURL, nil)
	if err != nil {
		return "", err
	}

	log.Debugf("checking upgrade version from %s", vURL)
	resp, err := v.HTTPClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("bad HTTP respose (status: %d) from %s",
			resp.StatusCode,
			vURL,
		)
	}

	decoder := yaml.NewDecoder(resp.Body)
	err = decoder.Decode(&targetVersion)
	if err != nil {
		return "", err
	}

	if v.O.Test {
		return targetVersion.Test, nil
	}
	return targetVersion.Production, nil
}

func (v upgradeCommand) Download(URL string, f *os.File, showProgress bool) error {
	var (
		done = make(chan int)
		wg   sync.WaitGroup
	)

	log.Debugf("will download %s to %s", URL, f.Name())

	client := v.HTTPClient()
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	defer f.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("cannot access %s (status: %d)", URL, resp.StatusCode)
	}

	if showProgress {
		contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
		if err != nil {
			log.Debugf("fail to get content-length: %s", err)
			contentLength = 0
		}

		wg.Add(1)
		go func(fullpath string, total int) {
			var (
				maxWidth = 60
				percent  float64
				stop     bool
				name     = filepath.Base(URL)
			)

			defer wg.Done()

			if total == 0 {
				return
			}

			for {
				select {
				case <-done:
					stop = true
				default:
				}

				if !stop {
					fi, err := os.Stat(fullpath)
					if err != nil {
						stop = true
						break
					}
					size := fi.Size()
					if size == 0 {
						size = 1
					}
					percent = float64(size) / float64(total) * 100
				} else {
					percent = 100
				}

				fmt.Printf("Download %s: %s %3.0f%%\r",
					name,
					strings.Repeat("#", int(percent/100*float64(maxWidth))),
					percent)

				if stop {
					break
				}

				time.Sleep(time.Second)
			}
			fmt.Printf("\n")
		}(f.Name(), contentLength)
	}

	_, err = io.Copy(f, resp.Body)

	if showProgress {
		done <- 1
		wg.Wait()

	}
	return err
}

func (v upgradeCommand) verifyChecksum(source, checksum string) error {
	var (
		err error
	)

	checksumFile, err := os.Open(checksum)
	if err != nil {
		return err
	}
	defer checksumFile.Close()

	expectChecksum := make([]byte, 64)
	if _, err := io.ReadFull(checksumFile, expectChecksum); err != nil {
		return fmt.Errorf("fail to read checksum file: %s", err)
	}
	log.Debugf("expect checksum: %s", expectChecksum)

	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actualChecksum := fmt.Sprintf("%x", h.Sum(nil))
	log.Debugf("actual checksum: %s", actualChecksum)

	if string(expectChecksum) != actualChecksum {
		return fmt.Errorf("bad checksum. %s != %s", expectChecksum, actualChecksum)
	}
	return nil
}

func (v upgradeCommand) verifySignature(source, sig string) error {
	var (
		err     error
		keyring openpgp.EntityList
	)

	log.Debug("validating signature")
	for _, buf := range config.PGPKeyRing {
		r := strings.NewReader(buf)
		keys, err := openpgp.ReadArmoredKeyRing(r)
		if err != nil {
			return fmt.Errorf("verify failed, cannot load pubkeys")
		}
		for _, key := range keys {
			for _, id := range key.Identities {
				log.Debugf("loaded pubkey for %s", id.Name)
			}
		}
		keyring = append(keyring, keys...)
	}

	sigFile, err := os.Open(sig)
	if err != nil {
		return err
	}
	defer sigFile.Close()

	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err != nil {
		return fmt.Errorf("verify failed, cannot read keyring: %s", err)
	}

	_, err = openpgp.CheckArmoredDetachedSignature(keyring, sourceFile, sigFile)
	if err != nil {
		return fmt.Errorf("fail to check pgp signature: %s", err)
	}

	log.Debugf("validating ok, good signature")
	return nil
}

func (v upgradeCommand) Verify(binURL, binFile string) error {
	var (
		checksumURL  string
		signatureURL string
	)

	// Check if have a sha256 signature
	checksumURL = binURL + ".sha256"
	checksumFile, err := ioutil.TempFile("", "git-repo.sha256-")
	if err != nil {
		return err
	}
	defer os.Remove(checksumFile.Name())

	err = v.Download(checksumURL, checksumFile, false)
	if err != nil {
		return fmt.Errorf("cannot find sha256 checksum: %s", checksumURL)
	}

	err = v.verifyChecksum(binFile, checksumFile.Name())
	if err != nil {
		return err
	}

	// Check pgp signature
	signatureURL = binURL + ".sha256.gpg"
	signatureFile, err := ioutil.TempFile("", "git-repo.sha256.gpg-")
	if err != nil {
		return err
	}
	defer os.Remove(signatureFile.Name())

	if err = v.Download(signatureURL, signatureFile, false); err != nil {
		input := userInput(
			fmt.Sprintf("cannot find pgp signature, still want to install? (y/N)? "),
			"N")
		if !answerIsTrue(input) {
			return fmt.Errorf("cannot valiate package, abort")
		}
	} else if err = v.verifySignature(checksumFile.Name(), signatureFile.Name()); err != nil {
		input := userInput(
			fmt.Sprintf("invalid pgp signature, still want to install? (y/N)? "),
			"N")
		if !answerIsTrue(input) {
			return fmt.Errorf("invalid package, abort")
		}
	}
	return nil
}

func (v upgradeCommand) UpgradeVersion(target, targetVersion string) error {
	var (
		binURL string
		err    error
	)

	binURL = fmt.Sprintf("%s/%s/%s/%s/%s",
		v.O.URL,
		targetVersion,
		runtime.GOOS,
		runtime.GOARCH,
		"git-repo",
	)

	if cap.IsWindows() {
		binURL += ".exe"
	}

	binFile, err := ioutil.TempFile("", "git-repo-")
	if err != nil {
		return err
	}
	defer os.Remove(binFile.Name())

	err = v.Download(binURL, binFile, true)
	if err != nil {
		return fmt.Errorf("fail to download %s: %s", binURL, err)
	}

	err = v.Verify(binURL, binFile.Name())
	if err != nil {
		return err
	}

	if config.IsDryRun() {
		log.Notef("will upgrade git-repo from version %s to %s, from file %s",
			version.Version,
			targetVersion,
			binFile.Name())
		return nil
	}

	err = v.InstallImage(binFile.Name(), target)
	if err != nil {
		return err
	}

	log.Notef("successfully upgrade git-repo from %s to %s",
		version.Version,
		targetVersion,
	)
	return nil
}

func (v upgradeCommand) InstallImage(bin, target string) error {
	var (
		lockFile string
	)

	if cap.IsWindows() {
		if strings.HasSuffix(strings.ToLower(target), ".exe") {
			target = target[0 : len(target)-4]
		}
		lockFile = target + "-new" + ".exe"
		target = target + ".exe"
	} else {
		lockFile = target + "-new"
	}

	in, err := os.Open(bin)
	if err != nil {
		return fmt.Errorf("cannot open src file while copying: %s", err)
	}
	defer in.Close()

	out, err := os.OpenFile(lockFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Debugf("no permission to write file '%s', will write to tmpfile instead",
			lockFile)
		out, err = ioutil.TempFile("", fmt.Sprintf("git-repo-%s-", v.O.Version))
		if err != nil {
			return err
		}
		lockFile = out.Name()
	}

	log.Debugf("first, install %s to %s", in.Name(), out.Name())
	_, err = io.Copy(out, in)
	if err != nil {
		out.Close()
		return fmt.Errorf("fail to copy from %s to %s", in.Name(), out.Name())
	}
	out.Close()

	os.Chmod(out.Name(), 0755)

	log.Debugf("at last, move %s to %s", out.Name(), target)
	err = os.Rename(out.Name(), target)
	if err != nil {
		box := format.NewMessageBox(78)
		box.Add("Fail to upgrade. Please copy")
		box.Add(fmt.Sprintf("        %s", out.Name()))
		box.Add("to")
		box.Add(fmt.Sprintf("        %s", target))
		box.Add("by hands")
		box.Draw(os.Stderr)
		return fmt.Errorf("upgrade failed")
	}

	return nil
}

func (v upgradeCommand) Execute(args []string) error {
	var (
		uv          string
		err         error
		mainProgram string
	)

	mainProgram, err = os.Executable()
	if err != nil {
		return err
	}
	if linkProgram, err := os.Readlink(mainProgram); err == nil {
		mainProgram = linkProgram
	}
	log.Debugf("program location: %s", mainProgram)

	if v.O.URL == "" {
		return errors.New("empty upgrade URL")
	}
	u, err := url.Parse(v.O.URL)
	if err != nil {
		return err
	}
	if u != nil && u.Host == "" {
		u, err = url.Parse("https://" + v.O.URL)
		if err != nil {
			return err
		}
		if u != nil && u.Host == "" {
			return fmt.Errorf("bad url: %s", v.O.URL)
		}
	}
	v.O.URL = u.String()
	if strings.HasSuffix(v.O.URL, "/") {
		v.O.URL = strings.TrimSuffix(v.O.URL, "/")
	}

	if v.O.Version == "" {
		uv, err = v.GetUpgradeVersion()
		if err != nil {
			return err
		}
	} else {
		uv = v.O.Version
	}

	if version.CompareVersion(uv, version.Version) <= 0 {
		if v.O.Version != "" {
			log.Warnf("will downgrade version from %s to %s", version.Version, v.O.Version)
		} else {
			log.Notef("current version (%s) is uptodate", version.Version)
			v.O.Version = uv
			return nil
		}
	} else {
		log.Debugf("compare versions: %s > %s", uv, version.Version)
	}

	if v.O.Version == "" {
		v.O.Version = uv
	}
	return v.UpgradeVersion(mainProgram, uv)
}

var upgradeCmd = upgradeCommand{}

func init() {
	rootCmd.AddCommand(upgradeCmd.Command())
}
