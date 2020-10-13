package helper

import (
	"errors"
	"net/url"

	"github.com/jiangxin/goconfig"
)

// GetProxyFromGitConfig returns http proxy url through http.proxy
func GetProxyFromGitConfig() (*url.URL, error) {
	var (
		err         error
		proxyRawURL string
		proxyURL    *url.URL
	)

	gitConfig, err := goconfig.LoadAll("")
	if err != nil {
		return nil, err
	}

	proxyRawURL = gitConfig.Get("http.proxy")
	if proxyRawURL == "" {
		return nil, errors.New("http.proxy is not set")
	}

	proxyURL, err = url.Parse(proxyRawURL)
	if err != nil {
		return nil, err
	}

	return proxyURL, nil
}
