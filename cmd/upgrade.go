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
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"errors"
	"fmt"
	"hash"
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

	"github.com/alibaba/git-repo-go/cap"
	"github.com/alibaba/git-repo-go/config"
	"github.com/alibaba/git-repo-go/file"
	"github.com/alibaba/git-repo-go/format"
	"github.com/alibaba/git-repo-go/helper"
	"github.com/alibaba/git-repo-go/version"
	log "github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/openpgp"
	"gopkg.in/yaml.v2"
)

const (
	// defaultUpgradeInfoURL is where to download version.yml from.
	defaultUpgradeInfoURL = "http://git-repo.info/download/version.yml"
	// defaultDownloadURL is where to download package file.
	defaultDownloadURL = "https://github.com/alibaba/git-repo-go/releases/download/v<version>/git-repo-<version>-<os>-<arch>.<ext>"
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

type upgradeInfo struct {
	Production string `yaml:"production"`
	Test       string `yaml:"test"`
	URLPattern string `yaml:"url"`
}

func (v *upgradeInfo) Version(isProduction bool) string {
	if isProduction {
		return v.Production
	}
	return v.Test
}

func (v *upgradeInfo) URLs(isProduction bool) []string {
	var urls []string

	url := v.URLPattern
	if url == "" {
		url = defaultDownloadURL
	}

	os := runtime.GOOS
	switch os {
	case "darwin":
		os = "macOS"
	case "linux":
		os = "Linux"
	case "windows":
		os = "Windows"
	}

	arch := runtime.GOARCH
	switch arch {
	case "386":
		arch = "32"
	case "amd64":
		arch = "64"
	}

	url = strings.ReplaceAll(url, "<version>", v.Version(isProduction))
	url = strings.ReplaceAll(url, "<os>", os)
	url = strings.ReplaceAll(url, "<arch>", arch)
	if strings.Contains(url, ".<ext>") {
		if os == "Windows" {
			urls = append(urls, strings.ReplaceAll(url, "<ext>", "zip"))
		} else {
			urls = append(urls,
				strings.ReplaceAll(url, "<ext>", "tar.gz"),
				strings.ReplaceAll(url, "<ext>", "tar.bz2"),
			)
		}
	} else {
		urls = append(urls, url)
	}

	return urls
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
		defaultUpgradeInfoURL,
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
		Proxy:                 http.ProxyFromEnvironment,
	}

	// http.proxy overrides env $HTTP_PROXY, $HTTPS_PROXY and $NO_PROXY (or the lowercase versions thereof).
	proxyURL, err := helper.GetProxyFromGitConfig()
	if err != nil {
		log.Debugf("fail to get proxy from git config: %s", err)
	} else {
		tr.Proxy = http.ProxyURL(proxyURL)
	}

	v.httpClient = &http.Client{Transport: tr}
	return v.httpClient
}

func (v upgradeCommand) GetUpgradeInfo() (*upgradeInfo, error) {
	var (
		info = upgradeInfo{}
	)

	infoURL := v.O.URL
	if strings.HasSuffix(v.O.URL, "/") {
		infoURL = infoURL + "version.yml"
	}

	req, err := http.NewRequest("GET", infoURL, nil)
	if err != nil {
		return nil, err
	}

	log.Debugf("checking upgrade version from %s", infoURL)
	resp, err := v.HTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot fetch upgrade info from %s, bad status: %d",
			infoURL,
			resp.StatusCode,
		)
	}

	decoder := yaml.NewDecoder(resp.Body)
	err = decoder.Decode(&info)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("fail to parse upgrade info from %s", infoURL)
	}
	return &info, nil
}

func (v upgradeCommand) Download(URL string, dir string, showProgress bool) (string, error) {
	var (
		done = make(chan int, 1)
		wg   sync.WaitGroup
	)

	fileName := filepath.Join(dir, filepath.Base(URL))
	log.Debugf("will download %s to %s", URL, fileName)

	client := v.HTTPClient()
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot access %s (status: %d)", URL, resp.StatusCode)
	}

	f, err := file.New(fileName).OpenCreateRewrite()
	if err != nil {
		return "", err
	}
	defer f.Close()

	if showProgress {
		contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
		if err != nil {
			log.Debugf("fail to get content-length: %s", err)
			contentLength = 0
		} else {
			log.Debugf("content-length for %s: %d", URL, contentLength)
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
		}(fileName, contentLength)
	}

	_, err = io.Copy(f, resp.Body)
	if showProgress {
		done <- 1
		wg.Wait()

	}
	return fileName, err
}

func (v upgradeCommand) verifyChecksum(source, checksum string) error {
	var (
		err error
		h   hash.Hash
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

	switch filepath.Ext(checksum) {
	case ".sha1":
		h = sha1.New()
	case ".sha512":
		h = sha512.New()
	case ".sha256":
		h = sha256.New()
	default:
		return fmt.Errorf("unknown checksum method for: %s", checksum)
	}
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

func (v upgradeCommand) ExtractPackage(pkgFile, dir string) (binFile, shaFile, gpgFile string, err error) {
	if strings.HasSuffix(pkgFile, ".zip") {
		return v.ExtractZip(pkgFile, dir)
	}
	return v.ExtractTar(pkgFile, dir)
}

func (v upgradeCommand) ExtractZip(pkgFile, dir string) (binFile, shaFile, gpgFile string, err error) {
	var (
		in  io.ReadCloser
		out *os.File
	)

	r, err := zip.OpenReader(pkgFile)
	if err != nil {
		err = fmt.Errorf("cannot open %s: %s", pkgFile, err)
		return
	}
	defer r.Close()

	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		fullpath := filepath.Join(dir, baseName)
		if baseName == "git-repo" || baseName == "git-repo.exe" {
			binFile = fullpath
		} else if strings.HasSuffix(baseName, ".sha256") ||
			strings.HasSuffix(baseName, ".sha512") ||
			strings.HasSuffix(baseName, ".sha1") {
			shaFile = fullpath
		} else if strings.HasSuffix(baseName, ".gpg") ||
			strings.HasSuffix(baseName, ".asc") {
			gpgFile = fullpath
		} else {
			log.Warningf("unknown file in package: %s", baseName)
			continue
		}
		in, err = f.Open()
		if err != nil {
			err = fmt.Errorf("extract error, cannot open %s", f.Name)
			return
		}
		out, err = file.New(fullpath).OpenCreateRewrite()
		if err == nil {
			_, err = io.Copy(out, in)
			if err != nil {
				err = fmt.Errorf("fail to write %s: %s", fullpath, err)
			}
			out.Close()
		} else {
			err = fmt.Errorf("cannot open %s to write: %s", fullpath, err)
		}
		in.Close()
		if err != nil {
			return
		}
	}
	return
}

func (v upgradeCommand) ExtractTar(pkgFile, dir string) (binFile, shaFile, gpgFile string, err error) {
	var reader io.Reader

	pkgf, err := os.Open(pkgFile)
	if err != nil {
		return
	}
	defer pkgf.Close()
	switch filepath.Ext(pkgFile) {
	case ".gz", ".tgz":
		reader, err = gzip.NewReader(pkgf)
	case ".bzip2", ".bz2":
		reader = bzip2.NewReader(pkgf)
	default:
		err = fmt.Errorf("unknown compress method for %s", filepath.Base(pkgFile))
	}

	if err != nil {
		return
	}

	tarReader := tar.NewReader(reader)
	for {
		var (
			header *tar.Header
			f      *os.File
		)

		header, err = tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return
		}

		name := header.Name
		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			break
		default:
			err = fmt.Errorf("unkown type \"%c\" for %s",
				header.Typeflag,
				name,
			)
			return
		}

		baseName := filepath.Base(name)
		if baseName == "git-repo" || baseName == "git-repo.exe" {
			binFile = filepath.Join(dir, baseName)
		} else if strings.HasSuffix(baseName, ".sha256") ||
			strings.HasSuffix(baseName, ".sha512") ||
			strings.HasSuffix(baseName, ".sha1") {
			shaFile = filepath.Join(dir, baseName)
		} else if strings.HasSuffix(baseName, ".gpg") ||
			strings.HasSuffix(baseName, ".asc") {
			gpgFile = filepath.Join(dir, baseName)
		} else {
			log.Warningf("unknown file in package: %s", baseName)
			continue
		}
		f, err = file.New(filepath.Join(dir, baseName)).OpenCreateRewrite()
		if err != nil {
			return
		}
		io.Copy(f, tarReader)
		f.Close()
	}

	return
}

func (v upgradeCommand) ExtractAndVerify(pkgFile, dir string) (string, error) {
	var (
		binFile, shaFile, gpgFile string
		err                       error
	)

	binFile, shaFile, gpgFile, err = v.ExtractPackage(pkgFile, dir)
	if binFile == "" {
		return "", fmt.Errorf("cannot find git-repo in package")
	}
	if shaFile == "" {
		return "", fmt.Errorf("cannot find checksum in package")
	}

	err = v.verifyChecksum(binFile, shaFile)
	if err != nil {
		return "", err
	}

	if gpgFile == "" {
		input := userInput(
			fmt.Sprintf("cannot find pgp signature, still want to install? (y/N)? "),
			"N")
		if !answerIsTrue(input) {
			return "", fmt.Errorf("cannot valiate package, abort")
		}
	} else if err = v.verifySignature(shaFile, gpgFile); err != nil {
		input := userInput(
			fmt.Sprintf("invalid pgp signature, still want to install? (y/N)? "),
			"N")
		if !answerIsTrue(input) {
			return "", fmt.Errorf("invalid package, abort")
		}
	}

	return binFile, nil
}

func (v upgradeCommand) UpgradeVersion(target string, info *upgradeInfo) error {
	var (
		downloadURL string
		pkgFile     string
		err         error
		newVersion  = info.Version(!v.O.Test)
	)

	tmpDir, err := ioutil.TempDir("", "git-repo-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	for _, downloadURL = range info.URLs(!v.O.Test) {
		pkgFile, err = v.Download(downloadURL, tmpDir, true)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("fail to download %s: %s", downloadURL, err)
	} else if downloadURL == "" {
		return fmt.Errorf("unknown download URL")
	}

	binFile, err := v.ExtractAndVerify(pkgFile, tmpDir)
	if err != nil {
		return err
	}

	if config.IsDryRun() {
		log.Notef("will upgrade git-repo from version %s to %s, from file %s",
			version.Version,
			newVersion,
			binFile)
		return nil
	}

	err = v.InstallImage(binFile, target)
	if err != nil {
		return err
	}

	log.Notef("successfully upgrade git-repo from %s to %s",
		version.Version,
		newVersion,
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

	out, err := file.New(lockFile).SetExecutable().OpenCreateRewrite()
	if err != nil {
		log.Debugf("no permission to write file '%s', will write to tmpfile instead",
			lockFile)
		tmpDir, err := ioutil.TempDir("", "git-repo-")
		if err != nil {
			return err
		}
		lockFile = filepath.Join(tmpDir, filepath.Base(target))
		out, err = file.New(lockFile).SetExecutable().OpenCreateRewrite()
		if err != nil {
			return err
		}
	}

	log.Debugf("first, install %s to %s", in.Name(), lockFile)
	_, err = io.Copy(out, in)
	if err != nil {
		out.Close()
		return fmt.Errorf("fail to copy from %s to %s", bin, lockFile)
	}
	out.Close()

	log.Debugf("at last, move %s to %s", out.Name(), target)
	err = os.Rename(lockFile, target)
	if err != nil {
		box := format.NewMessageBox(78)
		box.Add("ERROR: fail to upgrade. Please copy")
		box.Add(fmt.Sprintf("        %s", lockFile))
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
		mainProgram string
		info        *upgradeInfo
		newVersion  string
		err         error
	)

	if v.O.Test && v.O.Version != "" {
		return fmt.Errorf("cannot use --test and --version together")
	}

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

	if v.O.Version == "" {
		info, err = v.GetUpgradeInfo()
		if err != nil {
			return err
		}
	} else {
		info = &upgradeInfo{
			Production: v.O.Version,
		}
	}

	newVersion = info.Version(!v.O.Test)
	if version.CompareVersion(newVersion, version.Version) <= 0 {
		if v.O.Version != "" {
			log.Warnf("will downgrade version from %s to %s", version.Version, newVersion)
		} else {
			log.Notef("current version (%s) is uptodate", version.Version)
			return nil
		}
	} else {
		log.Debugf("compare versions: %s > %s", newVersion, version.Version)
	}

	return v.UpgradeVersion(mainProgram, info)
}

var upgradeCmd = upgradeCommand{}

func init() {
	rootCmd.AddCommand(upgradeCmd.Command())
}
