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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"code.alibaba-inc.com/force/git-repo/cap"
	"code.alibaba-inc.com/force/git-repo/config"
	"code.alibaba-inc.com/force/git-repo/format"
	"code.alibaba-inc.com/force/git-repo/path"
	"code.alibaba-inc.com/force/git-repo/versions"
	"github.com/jiangxin/multi-log"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/openpgp"
	"gopkg.in/yaml.v2"
)

const (
	// DefaultUpgradeURL indicates where to download git-repo new package
	DefaultUpgradeURL = "https://git-repo.oss-cn-zhangjiakou.aliyuncs.com"
)

type upgradeCommand struct {
	cmd        *cobra.Command
	httpClient *http.Client

	O struct {
		URL     string
		Test    bool
		Version string
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
			return v.runE(args)
		},
	}

	v.cmd.Flags().StringVarP(&v.O.URL,
		"url",
		"u",
		DefaultUpgradeURL,
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

	return v.cmd
}

func (v *upgradeCommand) HTTPClient() *http.Client {
	if v.httpClient != nil {
		return v.httpClient
	}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}

	v.httpClient = &http.Client{Transport: tr}
	return v.httpClient
}

func (v upgradeCommand) GetUpgradeVersion() (string, error) {
	var (
		version = upgradeVersion{}
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
	err = decoder.Decode(&version)
	if err != nil {
		return "", err
	}

	if v.O.Test {
		return version.Test, nil
	}
	return version.Production, nil
}

func (v upgradeCommand) Download(URL string, f *os.File) error {
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

	_, err = io.Copy(f, resp.Body)
	return err
}

func (v upgradeCommand) Verify(bin, sig string) error {
	var (
		err     error
		keyring openpgp.EntityList
	)

	log.Debugf("validating signature '%s' on '%s'", sig, bin)
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

	binFile, err := os.Open(bin)
	if err != nil {
		return err
	}
	defer binFile.Close()

	if err != nil {
		return fmt.Errorf("verify failed, cannot read keyring: %s", err)
	}

	_, err = openpgp.CheckArmoredDetachedSignature(keyring, binFile, sigFile)
	if err != nil {
		return fmt.Errorf("fail to check pgp signature: %s", err)
	}

	log.Debugf("validating ok, good signature")
	return nil
}

func (v upgradeCommand) UpgradeVersion(target, version string) error {
	var (
		binURL string
		sigURL string
		err    error
	)

	binURL = fmt.Sprintf("%s/%s/%s/%s/%s",
		v.O.URL,
		version,
		runtime.GOOS,
		runtime.GOARCH,
		"git-repo",
	)
	if cap.IsWindows() {
		binURL += ".exe"
	}
	sigURL = binURL + ".asc"

	binFile, err := ioutil.TempFile("", "git-repo-")
	if err != nil {
		return err
	}
	defer os.Remove(binFile.Name())

	sigFile, err := ioutil.TempFile("", "git-repo.asc-")
	if err != nil {
		return err
	}
	defer os.Remove(sigFile.Name())

	// Downlaod binURL and sigURL
	err = v.Download(binURL, binFile)
	if err != nil {
		return fmt.Errorf("cannot download package: %s", err)
	}

	err = v.Download(sigURL, sigFile)
	if err != nil {
		return fmt.Errorf("cannot download pgp signature: %s", err)
	}

	err = v.Verify(binFile.Name(), sigFile.Name())
	if err != nil {
		return err
	}

	if config.IsDryRun() {
		log.Notef("will upgrade git-repo from version %s to %s, from file %s",
			versions.Version,
			version,
			binFile.Name())
		return nil
	}

	err = v.InstallImage(binFile.Name(), target)
	if err != nil {
		return err
	}

	log.Notef("successfully upgrade git-repo from %s to %s",
		versions.Version,
		version,
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

func (v upgradeCommand) runE(args []string) error {
	var (
		uv          string
		err         error
		mainProgram string
	)

	mainProgram, err = path.Abs(os.Args[0])
	if err != nil {
		return err
	}
	if linkProgram, err := os.Readlink(mainProgram); err == nil {
		mainProgram = linkProgram
	}

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

	if versions.CompareVersion(uv, versions.Version) <= 0 {
		if v.O.Version != "" {
			log.Warnf("will downgrade version from %s to %s", versions.Version, v.O.Version)
		} else {
			log.Notef("current version (%s) is uptodate", versions.Version)
			v.O.Version = uv
			return nil
		}
	} else {
		log.Debugf("compare versions: %s > %s", uv, versions.Version)
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
