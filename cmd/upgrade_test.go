package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/aliyun/git-repo-go/file"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestUpgradeInfoURLs(t *testing.T) {
	var (
		assert   = assert.New(t)
		expect   []string
		osType   string
		archType string
		uinfo    upgradeInfo
	)

	switch runtime.GOOS {
	case "darwin":
		osType = "macOS"
	case "windows":
		osType = "Windows"
	case "linux":
		osType = "Linux"
	}
	switch runtime.GOARCH {
	case "386":
		archType = "32"
	case "amd64":
		archType = "64"
	}

	uinfo = upgradeInfo{
		Production: "0.9.1",
		Test:       "1.0.0.rc1",
	}
	assert.Equal("0.9.1", uinfo.Version(true), "production version")
	assert.Equal("1.0.0.rc1", uinfo.Version(false), "test version")
	if runtime.GOOS == "windows" {
		expect = []string{
			"https://github.com/aliyun/git-repo-go/releases/download/v0.9.1/git-repo-0.9.1-" + osType + "-" + archType + ".zip",
		}

	} else {
		expect = []string{
			"https://github.com/aliyun/git-repo-go/releases/download/v0.9.1/git-repo-0.9.1-" + osType + "-" + archType + ".tar.gz",
			"https://github.com/aliyun/git-repo-go/releases/download/v0.9.1/git-repo-0.9.1-" + osType + "-" + archType + ".tar.bz2",
		}
	}
	assert.Equal(expect, uinfo.URLs(true))

	if runtime.GOOS == "windows" {
		expect = []string{
			"https://github.com/aliyun/git-repo-go/releases/download/v1.0.0.rc1/git-repo-1.0.0.rc1-" + osType + "-" + archType + ".zip",
		}

	} else {
		expect = []string{
			"https://github.com/aliyun/git-repo-go/releases/download/v1.0.0.rc1/git-repo-1.0.0.rc1-" + osType + "-" + archType + ".tar.gz",
			"https://github.com/aliyun/git-repo-go/releases/download/v1.0.0.rc1/git-repo-1.0.0.rc1-" + osType + "-" + archType + ".tar.bz2",
		}
	}
	assert.Equal(expect, uinfo.URLs(false))

	uinfo = upgradeInfo{
		Production: "0.9.1",
		Test:       "1.0.0.rc1",
		URLPattern: "http://example.com/<version>/git-repo-<os>-<arch>.tgz",
	}
	expect = []string{
		"http://example.com/0.9.1/git-repo-" + osType + "-" + archType + ".tgz",
	}
	assert.Equal(expect, uinfo.URLs(true))
	expect = []string{
		"http://example.com/1.0.0.rc1/git-repo-" + osType + "-" + archType + ".tgz",
	}
	assert.Equal(expect, uinfo.URLs(false))
}

func TestGetUpgradeInfo(t *testing.T) {
	var (
		assert = assert.New(t)
	)

	defer gock.Off()

	gock.New("http://example.com").
		Get("/download/version.yml").
		Reply(404).
		BodyString("404 Not found")
	cmd := upgradeCommand{}
	cmd.O.URL = "http://example.com/download/version.yml"
	httpClient := cmd.HTTPClient()
	gock.InterceptClient(httpClient)
	uinfo, err := cmd.GetUpgradeInfo()
	assert.Equal("cannot fetch upgrade info from http://example.com/download/version.yml, bad status: 404", err.Error())
	assert.Nil(uinfo)

	gock.New("http://example.com").
		Get("/download/version.yml").
		Reply(200).
		BodyString("bad content")
	uinfo, err = cmd.GetUpgradeInfo()
	assert.Equal("fail to parse upgrade info from http://example.com/download/version.yml", err.Error())
	assert.Nil(uinfo)

	gock.New("http://example.com").
		Get("/download/version.yml").
		Reply(200).
		BodyString("production: 1.2.3\ntest: 2.0.5")
	uinfo, err = cmd.GetUpgradeInfo()
	assert.Nil(err)
	if assert.NotNil(uinfo) {
		assert.Equal("1.2.3", uinfo.Version(true))
		assert.Equal("2.0.5", uinfo.Version(false))
		assert.True(len(uinfo.URLs(true)) > 0)
	}

	gock.New("http://example.com").
		Get("/download/version.yml").
		Reply(200).
		BodyString("production: 0.1.2\ntest: 0.2.0\nurl: https://download.com/<version>.zip")
	uinfo, err = cmd.GetUpgradeInfo()
	assert.Nil(err)
	if assert.NotNil(uinfo) {
		assert.Equal("0.1.2", uinfo.Version(true))
		assert.Equal("0.2.0", uinfo.Version(false))
		assert.Equal("https://download.com/<version>.zip", uinfo.URLPattern)
		assert.Equal([]string{"https://download.com/0.1.2.zip"}, uinfo.URLs(true))
	}
}

func TestVerify(t *testing.T) {
	var (
		data     = "Hello, world.\n"
		checksum = "1ab1a2bb8502820a83881a5b66910b819121bafe336d76374637aa4ea7ba2616  hello.txt\n"
		sig      = `
-----BEGIN PGP SIGNATURE-----

iQJOBAABCgA4FiEEh/2pehIUeddqpt2hs/sBsKSIL5IFAlz3FUQaHHpoaXlvdS5q
eEBhbGliYWJhLWluYy5jb20ACgkQs/sBsKSIL5L6MRAAn8ncLGis6tAvOMwGS4R6
5vfCxtoWb8d7Fv9At6ANzo1M0vAeAfTfbXFqYV6oqzS89In1E9Km4Wi/i9Qr5yu8
u2L0gaDrK2n04K+NSGG61DcSTb2vyldpY5+I3Grbvw56G8WpnIKLmCwlF1GrhJWE
89G9p/WPATnkgaCtSbIaMSW0tW3PTl31XTUFqdQ2lutn+BDxY6ItjU5znESyCTC+
+3zFlIqqOo0iM4BCFLTwZk375EyxPj5WUJWWdbTMo7Bh5oJgNC7saaLEvCUQ9mi9
0ffQy3Zl1lykudxYlHJrNL+ym5VaB3l6rMe94v3cD6ObyPyy1jWh9r1QPOnQxCTq
lU3Y/MGynJfO1GBKE/bM8hY/t/Rt8unP9F/A7gMkvNDCjjVZ5qv6eKgvzkEMusjJ
XhRtsynEo4ybVzeHgBXopKqSR6A7kRemcq6DR7v/5iEoPzjIQgpd/P9VsYV+jTvU
V2jHrRxjRo1L7HGmpkwg1cdU0ewNjlrf+qtRFUPGGAW9ObSmsww6ggGCn/GXVZnG
43f0Tuo0BGrRR8s37P1TV0N0kCFPCqH/AxWA9UUr4cjTSiaOC2M1vNDAYyLWNDx9
qUDzwGflENTCedt+7lz7oLhAk3Oor0Gxk95u43ki9REJMIkS68CXDfe3t4GKUrKa
9GO21W1p3viN2eJhj76/oZU=
=K8cW
-----END PGP SIGNATURE-----
`

		err    error
		assert = assert.New(t)
	)

	binFile, err := ioutil.TempFile("", "test-git-repo.-")
	if err != nil {
		t.Fatalf("fail to create binfile: %s", err)
	}
	defer os.Remove(binFile.Name())
	binFile.WriteString(data)
	binFile.Close()

	checksumFile, err := file.New(binFile.Name() + ".sha256").OpenCreateRewrite()
	if err != nil {
		t.Fatalf("fail to create checksum file: %s", err)
	}
	defer os.Remove(checksumFile.Name())
	checksumFile.WriteString(checksum)
	checksumFile.Close()

	sigFile, err := file.New(binFile.Name() + ".sha256.gpg").OpenCreateRewrite()
	if err != nil {
		t.Fatalf("fail to create sig file: %s", err)
	}
	defer os.Remove(sigFile.Name())
	sigFile.WriteString(sig)
	sigFile.Close()

	cmd := upgradeCommand{}
	err = cmd.verifyChecksum(binFile.Name(), checksumFile.Name())
	assert.Nil(err)

	err = cmd.verifySignature(checksumFile.Name(), sigFile.Name())
	assert.Nil(err)
}

func TestDownload(t *testing.T) {
	var (
		err    error
		assert = assert.New(t)
		data   = "hello, world"
	)

	defer gock.Off()

	gock.New("http://example.com").
		Get("/download/filename").
		Reply(404).
		BodyString("404 Not found")

	cmd := upgradeCommand{}
	httpClient := cmd.HTTPClient()
	gock.InterceptClient(httpClient)
	file, err := cmd.Download("http://example.com/download/filename", "/tmp", false)
	if assert.NotNil(err) {
		assert.Equal("cannot access http://example.com/download/filename (status: 404)", err.Error())
	}
	assert.Equal("", file)

	gock.New("http://example.com").
		Get("/download/filename").
		Reply(200).
		AddHeader("Content-Length", strconv.Itoa(len(data))).
		BodyString(data)
	dir, err := ioutil.TempDir("", "git-repo-test-")
	if assert.Nil(err) {
		defer os.RemoveAll(dir)
		file, err = cmd.Download("http://example.com/download/filename", dir, false)
		assert.Nil(err)
		if assert.True(file != "") {
			buf, err := ioutil.ReadFile(file)
			if assert.Nil(err) {
				assert.Equal([]byte(data), buf)
			}
		}
	}
}

func TestExtract(t *testing.T) {
	var (
		assert = assert.New(t)
		data   = map[string][]byte{
			"tar.gz": []byte(`H4sIAGE3GF4AA+2WybKiSBSGa81TuLfrMiSZwKIWiMgoMsngjlFQBAQU9enLtqpvR0XXsKlb1R3ttzkRGZl5DvznP5HkC/FC4O/eFIIgGAgnj4jQIxIU/Sl+ZkICRCACAEiSE4KkaATeTeDblvWJUz9E3b2UXRnV20tZf2vffVuef+eez9/xGv8jkA/9t+Xwvsva5m1y3P8Hounv6E/CV/1JBO760yS460+8TTlf8j/Xv8iqqvljMjZdlWK/u5gnv5wv/f/SFxEF0cu23f7EHD/wPwlp9Op/GsG7/yED4NP/v4L3fzITJcWYmJI5cRTJ4N21LT7WMay01NWM52cCz9OLUhQLnGqzQllnaXpsB6ro8X7Wa46iQ2XBV3Ro3KxIltsmCqpz6sAjlomzIpaqMvTVQverU3iFu5gieGG7t/4+rFO8hePBmjuF22DQ3W7ou5R0VtwFy3VyVFxnFh4XmmptR3XEE+0iU66OLsyo2PUZdLLp0/cLWg0W/U3tEic9qo3GcOYi7DAvPAiR4eNp4OCUcKsPyDHb4jBLacscDBAnbrasKrKt1bDSlBwXWfmaJSfuhuvWatGlWHCg16XjNHArlHlTiV6tVFd7le20LibaPr51+GocVwyqe7THe6Usw6xxiiVvZ7lwaERsWgCvyy7jLnGC8xWwtRTYA2NTS+hU23x3SvbZNNisWKLO1IsbTYMjZ6ulxNazkulKr8EE3+0Tlu3YgV8F5jJmkoQfb8kSNMepLqr+kZKyzR76u5zPREfhRU8w/FPLJa3bEjh/xkBZi56RW8b5Jld2Y/bzpT0LyEW8VY3pAK7mgU1HA+z3nDUXEs3RdJiq84O898Fy0ZU+5l73aKAldVO0fXQqLE06GX1tHPy4MEbu6kaH/HbahG1rrjVj1Dq1sKkM4tHmrM8amSaw1mhWFxJoLFcJi76/MfNWTAQrbx0U2pS4dPWFdG+CUynEdF30/lUwjEVWVvKq1/eAoDFeYo/aer3Wpxdehb4qZlWgeE067+Or0yjdIPbDUk8J6yrz7uiN0PUqa1OWiO01zzFHbCnNWe8mG6bW8F6ugXM+xau9XkJlbKOmmNXBYTxeffO0ku3FnFQrnkCuvr4Xu8FrWxQxT+Y2Yn8zTRGGnOpYhjDbeGH/AftwQrsIezhGNOZf89Hvdvm3+er8/8k5fvT+I0nwOv8p8Jj/9xfgc/7/ClgI8pwDDKIiAqVpnDMUldBZnHFpmiKUsjkCaRplEcfkkCITkCUJRaQRk3AMIihiMvmrd/7FTf7kyZMnT/7BR+RIk84AFAAA`),

			"tar.bz2": []byte(`QlpoOTFBWSZTWYqNiAIAANH/uP6QBIBAD//iP///8H////ABAAQAAQBACFADXnTy9y9b2u0iIoKfpTZT02pPTTBoSPUe0mKYyGTJo1PKeoaAABqqn6aap+0TVP1PTE9KYRmmTTDRNNojJNqn6aTUABiBqkI9TQ9B6BT2mUekjNTwmp5T09U8piZDQaDT1PU0Bqp4SaegaDUynqNHtUPUyMT1G1PFPUMjJoAAASkVPaJNqeU9GU02iZG1MgDQZBmKBoDT1GhtTMEwTx+cWxaUa9cKUWpFOiADLNobAPcKgNoIRgsEg73l709F+839SuPVc/3E/ABbCyCgeRLDGJYsNKWOJHDl7KQ6+gnmZe3Q7hAUKN6tQDQDYDNsWdOrbaJ+IfPzyOJh9NbQUXEbAN3FiQ0KRoAL4GMFBN4Zn2gbAUDoWkOJQNJLgWg/hUsMueve04NNN2Wq5UrBjbZMqZSvWuKL7zMq+XNTe0BDKETcU1hPBcBqbrjAbxGBYQCpQU4Gv+Lx02txx2H9NOWKB8Ce5GkO3HGIX8jU36uJeFDZHHc5i20FwQC5P5UfZRenZPcknhhC6uDbCIFtVomJqNBbhtsddXTlqQ11u5OBjAB+2rnZdRGjk4vK6N/flnxoPZri4mfXNmoN8YlvQsOvXvlmqqs5K1cWSaOxMkNTYK6UN392edqCL70JODMHo3kOtPMxW4dYsvsTRJmQU43PZbTDpiPKtYcEhLr+2G9clX7jRQRSw+GJ1tVfjvi4C5fKzdpkNw56a3JcNHA9qAE3EVXszlaU0bulJ1LtSvcmWSKNxufNPNksVyyOZyBUM6If8JBR6QHP2/2O3bTi6redTUJmxGwq+VWtBmAeJLLJz4KjQdFG/G9e2BlMh5mHZMgy7+wcqnyK2Y8TFKLAWt7+CTXrTeQWNbZQjZdlSMbjLnLCQw0nnieECmfyjK9mGOnoYSXkZvhnSy1FPhZq2D49yUEkexbKCHLyN3mJL1ebrPRjTJwyZOw99aLde288VPlXo3I3xsOw5RY1K810RSyVURJpKWc8hALoqOIzpk10+CWgBEYFJgFanmrHqEuS1NCvnUbpPU6rO/B2pRgmNlu0EWqfbLdisQMb5cOGIVFDa2mqUs0enFPXbJmngjp6DAtAHt5WmBQuzt1OAFrPTychXbNFcgJ/yq+Bo5zF5GXoRqFHybt3LlqykxRLaZIQJg6gBscK9wA++AHz47a6QCMTCc1CqfdWPkpa3QM9kqS5Iac/ASjVrxpbqR30UQYFgDaabk0S49BUgDf2w8EQSYOAl9aZ76yLhFV5Nhrqc5TqkjJUbXewBG8/p7Hlj2p9dqDGYDDsSzvDVXD+gHb8LEzxbtzDw8EIgz5kYfgQWvEfZkdfYAwCuLCcDF5nnx17FSoABH1nsS5cdt8Tpv8Ba/0UlZ0rmkRkq4I798PX5z2zN58Dz+xUYBVx2OvNk4qVVrnxOpnARu5KzyMAz/F3JFOFCQio2IAg`),

			"zip": []byte(`UEsDBBQAAAAAAPiDKlBTdCT0DQAAAA0AAAAOAAAAMS4wLjAvZ2l0LXJlcG9oZWxsbywgd29ybGQKUEsDBBQAAAAIAA+EKlDuOPwbRAAAAEsAAAAVAAAAMS4wLjAvZ2l0LXJlcG8uc2hhMjU2BcHRDYAwCAXAf6foAiZILZRxkAfGL41x/3g3R6+yrsJOAhylzLHnkQZABLOkA55uWoO36BnBBNcwFWJq7by+9c3nXn5QSwMEFAAAAAgAYYQqUFrGNyuqAgAAZgMAABkAAAAxLjAuMC9naXQtcmVwby5zaGEyNTYuZ3BnbZM5s6JAAIRzfsXm1isURCF4wYDDzXAMcpjJJcgpAyL8+n27ySbbYVd1Vwf9fX39SISKhn45ivMLawoC/tWDf+wviqpc3RYBECUAjnIFYUkzQ15q1zzLXsPElIQmIjGwZnKaDJpjjDb3rqpDf4+ad4a5F5VDsUyUpopDvTTDZo5X7pkweyA9avdf2GSAS9PRVZjjRzSZ/jiRMTtgW/hQhXlYNB+L8Us2dPex6AudGh+V8c3T57xoXvdmR9UJjz8Fg8GVZNPHFGcvvTfOgiPHIxXErXRHIZ1FmGakrWtP2BnKVsyOrjMhNkn93Gqaw9DpcWNoBQ15dc3TWdho07XlMaOi9nitMO65h1QVfQODTmtWz86fxpjsB5JsI20vi30+deRU00SrqjjvcWkBLy+ktofUrmSDMf8szxRH75XlOyXyprPHWBxuHsVzTut8F91sft/l+se/76KX4OmVwndidR6roKek0Ccpz4/8BOzIsZJzmoJlSy22f+1MqIcvRslvNRc+C5BDrAEYSCicByEd/GFPgzfFVh0MUOGi96Y2Xu+Qi+WJ0UFOHjraTezqtHy2ILauBfcipQY2TC7TL61ah6wlj1VI+Wt9mo6KfisHcp9L11BmRDrUhkmJFmH1722xzbd4GJyrgRZj1EuPyTn6fnubYq8e99SAevtzYA1eaCSZkO18GWAqucWAT7HHQMs3ZeXnBHMlJceuJOEqISTnVaPaxKzZ/ZECCv8yrterufsAnQt1mDeRFvTZhSQr7rVxgmSyzGzvrirwl2Dh/KBxb1V14okRYGehLOXCB5uKHKMHQWGw72JHN7VZcdoy3PtS7KJ2ea2hM9uqJ18OegP2J9+8/oy90Z0HIRWowg2SzXEgFws6dpEk3oKYfFPf8+l5p/4SA9Hlfxz9BlBLAQIUABQAAAAAAPiDKlBTdCT0DQAAAA0AAAAOAAAAAAAAAAEAIAAAAAAAAAAxLjAuMC9naXQtcmVwb1BLAQIUABQAAAAIAA+EKlDuOPwbRAAAAEsAAAAVAAAAAAAAAAEAIAAAADkAAAAxLjAuMC9naXQtcmVwby5zaGEyNTZQSwECFAAUAAAACABhhCpQWsY3K6oCAABmAwAAGQAAAAAAAAABACAAAACwAAAAMS4wLjAvZ2l0LXJlcG8uc2hhMjU2LmdwZ1BLBQYAAAAAAwADAMYAAACRAwAAAAA=`),
		}
	)

	cmd := upgradeCommand{}
	for key, buf := range data {
		dir, err := ioutil.TempDir("", "git-repo-test")
		if assert.Nil(err) {
			defer os.RemoveAll(dir)
			content, err := base64.StdEncoding.DecodeString(string(buf))
			if assert.Nil(err) {
				fileName := filepath.Join(dir, "file-1.0.0."+key)
				err = ioutil.WriteFile(fileName, content, 0644)
				if assert.Nil(err) {
					binFile, err := cmd.ExtractAndVerify(fileName, dir)
					if assert.Nil(err, fmt.Sprintf("ExractPackage: %s", fileName)) {
						actual, err := ioutil.ReadFile(binFile)
						assert.Nil(err)
						assert.Equal("hello, world\n", string(actual))
					}
				}
			}
		}
	}
}
