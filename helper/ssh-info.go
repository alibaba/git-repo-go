package helper

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/git-repo-go/config"
	"github.com/alibaba/git-repo-go/path"
	"github.com/jiangxin/goconfig"
	log "github.com/jiangxin/multi-log"
	"gopkg.in/h2non/gock.v1"
)

const (
	sshInfoCmdTimeout         = 3
	remoteCallTimeout         = 10
	sshInfoCacheDefaultExpire = 3600 * 12 // Seconds
	expireTimeLayout          = "2006-01-02 15:04:05"
)

var (
	sshInfoPattern = regexp.MustCompile(`^[\S]+ [0-9]+$`)
	httpClient     *http.Client
	internalCache  map[string]interface{}
)

// SSHInfo wraps host and port which ssh_info returned.
type SSHInfo struct {
	Host         string `json:"host,omitempty"`
	Port         int    `json:"port,omitempty"`
	ProtoType    string `json:"type,omitempty"`
	ProtoVersion int    `json:"version,omitempty"`
	User         string `json:"user,omitempty"`

	Expire int64 `json:"-"`
}

// ToJSON encodes ssh_info to JSON.
func (v SSHInfo) ToJSON() string {
	buf, err := json.Marshal(&v)
	if err != nil {
		log.Errorf("fail to marshal ssh_info: %s", err)
	}
	return string(buf)
}

// SSHInfoQuery wraps cache to accelerate query of ssh_info API.
type SSHInfoQuery struct {
	CacheFile string

	cfg     goconfig.GitConfig
	Changed bool
}

// GetSSHInfo queries ssh_info for address.
func (v SSHInfoQuery) GetSSHInfo(address string, useCache bool) (*SSHInfo, error) {
	var (
		err error
	)

	key := urlToKey(address)
	if key == "" {
		return nil, fmt.Errorf("bad address for review '%s'", address)
	}

	// Try internal cache
	if cache, ok := internalCache[key]; ok {
		switch cache.(type) {
		case error:
			return nil, cache.(error)
		case *SSHInfo:
			return cache.(*SSHInfo), nil
		default:
			log.Errorf("bad internal cache for SSHInfo: %#v", cache)
		}
	}

	// Try cache
	if v.CacheFile != "" && v.cfg != nil && useCache {
		data := v.cfg.Get(fmt.Sprintf(config.CfgManifestRemoteSSHInfo, key))
		if data != "" {
			expired := true
			expireStr := v.cfg.Get(fmt.Sprintf(config.CfgManifestRemoteExpire, key))
			if expireStr != "" {
				expireTm, err := time.Parse(expireTimeLayout, expireStr)
				if err == nil && expireTm.After(time.Now()) {
					expired = false
				}
			}
			if !expired {
				sshInfo := SSHInfo{}
				err = json.Unmarshal([]byte(data), &sshInfo)
				if err == nil && sshInfo.ProtoType != "" {
					log.Debugf("get ssh_info cache from '%s': '%s'", v.CacheFile, string(data))
					return &sshInfo, nil
				}
				log.Warnf("fail to parse ssh_info cache from '%s': '%s'", v.CacheFile, data)
			} else {
				log.Debugf("cache of ssh_info in '%s' is expired", v.CacheFile)
			}
		}
	}

	// Call ssh_info API
	sshInfo, err := querySSHInfo(address)
	if err != nil {
		// Update internal cache
		internalCache[key] = err
		return nil, err
	}
	// Update internal cache
	internalCache[key] = sshInfo
	log.Debugf("query ssh_info successfully: %#v", sshInfo)

	// Update Cache
	if v.CacheFile != "" && v.cfg != nil {
		path.SafeCreateParentDir(v.CacheFile)
		data := sshInfo.ToJSON()
		expireTime := time.Now().Add(time.Second * sshInfoCacheDefaultExpire).Format(expireTimeLayout)
		v.cfg.Set(fmt.Sprintf(config.CfgManifestRemoteSSHInfo, key), data)
		v.cfg.Set(fmt.Sprintf(config.CfgManifestRemoteExpire, key), expireTime)
		v.cfg.Save(v.CacheFile)
		log.Debugf("save cache file '%s', expire at '%s', data: '%s'", v.CacheFile, expireTime, data)
	}

	return sshInfo, nil
}

// NewSSHInfoQuery creates new query object. file is name of the cache.
func NewSSHInfoQuery(cacheFile string) *SSHInfoQuery {
	query := SSHInfoQuery{CacheFile: cacheFile}
	if cacheFile != "" {
		cfg, _ := goconfig.Load(cacheFile)
		if cfg == nil {
			cfg = goconfig.NewGitConfig()
		}
		query.cfg = cfg
	}
	return &query
}

// querySSHInfo queries ssh_info API and return SSHInfo object.
func querySSHInfo(address string) (*SSHInfo, error) {
	env := os.Getenv("REPO_HOST_PORT_INFO")
	if env != "" {
		log.Debugf("parse ssh_info from env: '%s'", env)
		return sshInfoFromString(env)
	}

	if strings.HasPrefix(address, "persistent-") {
		address = address[len("persistent-"):]
	}

	if address == "" {
		return &SSHInfo{}, nil
	}

	// Compatible with android repo.
	if strings.HasPrefix(address, "sso:") ||
		os.Getenv("REPO_IGNORE_SSH_INFO") != "" {
		log.Debug("REPO_IGNORE_SSH_INFO is defined, fallback to gerrit protocol")
		return &SSHInfo{ProtoType: ProtoTypeGerrit}, nil
	}

	url := config.ParseGitURL(address)
	if url == nil {
		sshInfo, err := querySSHInfo("https://" + address)
		if err != nil {
			sshInfo, err = querySSHInfo("http://" + address)
		}
		if err != nil {
			return nil, err
		}
		return sshInfo, nil
	}
	if url.IsSSH() {
		sshInfo, err := sshInfoFromCommand(url)
		if err != nil {
			return nil, err
		}
		return sshInfo, nil
	}
	sshInfo, err := sshInfoFromAPI(url)
	if err != nil {
		return nil, err
	}
	return sshInfo, nil
}

// sshInfoFromAPI queries ssh_info API and return SSHInfo object.
func sshInfoFromAPI(url *config.GitURL) (*SSHInfo, error) {
	var (
		err error
	)

	infoURL := url.GetRootURL() + "/ssh_info"

	// Mock ssh_info API
	if config.GetMockSSHInfoResponse() != "" || config.GetMockSSHInfoStatus() != 0 {
		mockStatus := config.GetMockSSHInfoStatus()
		if mockStatus == 0 {
			mockStatus = 200
		}
		mockResponse := config.GetMockSSHInfoResponse()
		gock.New(infoURL).
			Reply(mockStatus).
			BodyString(mockResponse)
	}

	log.Debugf("get ssh_info from API: %s", infoURL)
	req, err := http.NewRequest("GET", infoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("bad ssh_info access to '%s': %s", infoURL, err)
	}
	req.Header.Set("Accept", "application/json")

	client := getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bad ssh_info request to '%s': %s", infoURL, err)
	}
	defer resp.Body.Close()

	// Successful status code maybe 200, 201.
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("%d: bad ssh_info response of '%s'",
			resp.StatusCode,
			infoURL)
	}

	reader := bufio.NewReader(resp.Body)
	var buf bytes.Buffer
	for i := 0; i < 100; i++ {
		line, err := reader.ReadString('\n')
		buf.WriteString(line)
		if err != nil {
			break
		}
	}

	sshInfo, err := sshInfoFromString(buf.String())
	if err != nil {
		return nil, fmt.Errorf("fail to run ssh_info API on %s: %s",
			url.GetRootURL(),
			err)
	}
	return sshInfo, nil
}

// sshInfoFromCommand queries ssh_info ssh command and return the parsed SSHInfo object.
func sshInfoFromCommand(url *config.GitURL) (*SSHInfo, error) {
	var (
		sshInfo *SSHInfo
		err     error
		out     []byte
	)

	if url == nil || !url.IsSSH() {
		return nil, fmt.Errorf("bad protocol, ssh_info only apply for SSH")
	}

	cmdArgs, _ := NewSSHCmd().Command(url.UserHost(), url.Port, nil)
	cmdArgs = append(cmdArgs, "ssh_info")

	// Mock ssh_info API
	if config.GetMockSSHInfoResponse() != "" || config.GetMockSSHInfoStatus() != 0 {
		log.Notef("mock executing: %s", strings.Join(cmdArgs, " "))
		mockStatus := config.GetMockSSHInfoStatus()
		if mockStatus < 400 {
			mockStatus = 0
		}
		mockResponse := config.GetMockSSHInfoResponse()
		if mockStatus != 0 {
			err = fmt.Errorf("exit %d", mockStatus)
		} else {
			sshInfo, err = sshInfoFromString(mockResponse)
		}
	} else {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			sshInfoCmdTimeout*time.Second,
		)
		defer cancel()
		log.Debugf("get ssh_info from command: %s", strings.Join(cmdArgs, " "))
		out, err = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...).Output()
		if err != nil {
			err = fmt.Errorf("pipe ssh_info cmd failed: %s", err)
		} else {
			sshInfo, err = sshInfoFromString(string(out))
		}
	}

	if err != nil {
		// Gerrit's well known port: 29418
		if url.Port == 29418 {
			return &SSHInfo{
				Host:      url.Host,
				Port:      url.Port,
				ProtoType: ProtoTypeGerrit}, nil
		}

		log.Debug("fail to check ssh_info for SSH protocol, will check HTTP instead")
		return querySSHInfo(url.Host)
	}
	return sshInfo, nil
}

func sshInfoFromString(data string) (*SSHInfo, error) {
	var (
		sshInfo = SSHInfo{}
		err     error
	)

	data = strings.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("empty ssh_info")
	}

	// If `info` contains '<', we assume the server gave us some sort
	// of HTML response back, like maybe a login page.
	//
	// Assume HTTP if SSH is not enabled or ssh_info doesn't look right.
	if strings.HasPrefix(data, "<") {
		return nil, fmt.Errorf("ssh_info returns a normal HTML response")
	}

	if !strings.ContainsAny(data, "\n") {
		if data == "NOT_AVAILABLE" {
			sshInfo.ProtoType = ProtoTypeGerrit
			return &sshInfo, nil
		}
		if sshInfoPattern.MatchString(data) {
			items := strings.SplitN(data, " ", 2)
			if len(items) != 2 {
				return nil, fmt.Errorf("bad format: %s", data)
			}

			port, err := strconv.Atoi(items[1])
			if err != nil {
				return nil, fmt.Errorf("bad port number '%s': %s", items[1], err)
			}
			sshInfo.Port = port
			sshInfo.Host = items[0]
			sshInfo.ProtoType = ProtoTypeGerrit
			return &sshInfo, nil
		}
	}

	err = json.Unmarshal([]byte(data), &sshInfo)
	if err != nil {
		return nil, err
	}
	return &sshInfo, nil
}

func getHTTPClient() *http.Client {
	if httpClient != nil {
		return httpClient
	}

	skipSSLVerify := config.NoCertChecks()

	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   remoteCallTimeout * time.Second,
			KeepAlive: remoteCallTimeout * time.Second,
		}).DialContext,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: skipSSLVerify},
		TLSHandshakeTimeout:   remoteCallTimeout * time.Second,
		ResponseHeaderTimeout: remoteCallTimeout * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          10,
		IdleConnTimeout:       remoteCallTimeout * time.Second,
		DisableCompression:    true,
	}

	httpClient = &http.Client{Transport: tr}

	// Mock ssh_info API
	if config.GetMockSSHInfoResponse() != "" || config.GetMockSSHInfoStatus() != 0 {
		gock.InterceptClient(httpClient)
	}

	return httpClient
}

func urlToKey(address string) string {
	var (
		u   = config.ParseGitURL(address)
		key = ""
	)

	if u == nil {
		log.Debugf("fail to parse url: %s", address)
		return ""
	}

	if u.IsHTTP() {
		key = u.Proto + "://"
		key += u.Host
		if u.Port > 0 && u.Port != 80 && u.Port != 443 {
			key += fmt.Sprintf(":%d", u.Port)
		}
	} else if u.IsSSH() {
		key = u.Proto + "://"
		key += u.Host
		if u.Port > 0 && u.Port != 22 {
			key += fmt.Sprintf(":%d", u.Port)
		}
	}
	return key
}

func init() {
	internalCache = make(map[string]interface{})
}
