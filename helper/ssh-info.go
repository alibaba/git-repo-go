package helper

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.alibaba-inc.com/force/git-repo/config"
	"github.com/jiangxin/goconfig"
	log "github.com/jiangxin/multi-log"
	"gopkg.in/h2non/gock.v1"
)

const (
	remoteCallTimeout = 10
)

var (
	sshInfoPattern = regexp.MustCompile(`^[\S]+ [0-9]+$`)
	httpClient     *http.Client
)

// SSHInfo wraps host and port which ssh_info returned.
type SSHInfo struct {
	Host         string `json:"host,omitempty"`
	Port         int    `json:"port,omitempty"`
	ProtoType    string `json:"type,omitempty"`
	ProtoVersion int    `json:"version,omitempty"`
	Expire       int64  `json:"-"`
}

// TODO: save ssh_info to cache.
func SaveCache(sshInfo *SSHInfo, config goconfig.GitConfig, filename string) error {
	return nil
}

// TODO: save ssh_info to cache.
func LoadCache(remoteName string) (goconfig.GitConfig, error) {
	return nil, nil
}

// QuerySSHInfo queries ssh_info API and return SSHInfo object.
func QuerySSHInfo(address string) (*SSHInfo, error) {
	url := config.ParseGitURL(address)
	if url == nil {
		sshInfo, err := QuerySSHInfo("https://" + address)
		if err != nil {
			sshInfo, err = QuerySSHInfo("http://" + address)
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
		sshInfo = SSHInfo{}
		err     error
	)

	infoURL := url.GetReviewURL() + "/ssh_info"

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
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("fail to access ssh_info URL '%s': %s", infoURL, err)
	}

	if strings.HasPrefix(line, "NOT_AVAILABLE") {
		sshInfo.ProtoType = config.RemoteTypeGerrit
		return &sshInfo, nil
	}

	// If `info` contains '<', we assume the server gave us some sort
	// of HTML response back, like maybe a login page.
	//
	// Assume HTTP if SSH is not enabled or ssh_info doesn't look right.
	if strings.HasPrefix(line, "<") {
		/*
			log.Notef("get a normal html response, may be a bad config gerrit server?")
			sshInfo.ProtoType = config.RemoteTypeGerrit
			return &sshInfo, nil
		*/
		return nil, fmt.Errorf("ssh_info on '%s' has a normal HTML response", infoURL)
	}

	buf := bytes.NewBufferString(line)
	n := 0
	for {
		line, err = reader.ReadString('\n')
		buf.WriteString(line)
		if err != nil || n > 100 {
			break
		}
		n++
	}
	data := buf.String()
	data = strings.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("empty ssh_info on '%s'", infoURL)
	}
	if sshInfoPattern.MatchString(data) {
		items := strings.SplitN(data, " ", 2)
		if len(items) != 2 {
			return nil, fmt.Errorf("bad ssh_info response on '%s': %s", infoURL, data)
		}

		port, err := strconv.Atoi(items[1])
		if err != nil {
			return nil, fmt.Errorf("bad port number (%s) in ssh_info on '%s': %s", items[1], infoURL, err)
		}
		sshInfo.Port = port
		sshInfo.Host = items[0]
		sshInfo.ProtoType = config.RemoteTypeGerrit
	} else {
		err = json.Unmarshal([]byte(data), &sshInfo)
		if err != nil {
			return nil, fmt.Errorf("fail to parse ssh_info response on '%s': %s", infoURL, data)
		}
	}
	return &sshInfo, nil
}

// sshInfoFromCommand queries ssh_info ssh command and return the parsed SSHInfo object.
func sshInfoFromCommand(url *config.GitURL) (*SSHInfo, error) {
	if url == nil || !url.IsSSH() {
		return nil, fmt.Errorf("bad protocol, ssh_info only apply for SSH")
	}

	cmdArgs := []string{"ssh"}
	if url.User != "" {
		cmdArgs = append(cmdArgs, "-l", url.User)
	}
	if url.Port > 0 && url.Port != 22 {
		cmdArgs = append(cmdArgs, "-p", strconv.Itoa(url.Port))
	}
	cmdArgs = append(cmdArgs, url.Host, "ssh_info")

	log.Debugf("will execute: %s", strings.Join(cmdArgs, " "))
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("pipe ssh_info cmd failed: %s", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start ssh_info cmd failed: %s", err)
	}

	sshInfo, err := sshInfoFromReader(out)

	if err2 := cmd.Wait(); err2 != nil {
		return nil, fmt.Errorf("execute ssh_info cmd failed: %s", err2)
	}

	if err != nil {
		return nil, fmt.Errorf("fail to run ssh_info cmd on %s: %s",
			url.GetReviewURL(),
			err)
	}
	return sshInfo, nil
}

func sshInfoFromReader(r io.Reader) (*SSHInfo, error) {
	sshInfo := SSHInfo{}
	reader := bufio.NewReader(r)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}

	if strings.HasPrefix(line, "NOT_AVAILABLE") {
		sshInfo.ProtoType = config.RemoteTypeGerrit
		return &sshInfo, nil
	}

	// If `info` contains '<', we assume the server gave us some sort
	// of HTML response back, like maybe a login page.
	//
	// Assume HTTP if SSH is not enabled or ssh_info doesn't look right.
	if strings.HasPrefix(line, "<") {
		return nil, fmt.Errorf("ssh_info returns a normal HTML response")
	}

	buf := bytes.NewBufferString(line)
	n := 0
	for {
		line, err = reader.ReadString('\n')
		buf.WriteString(line)
		if err != nil || n > 100 {
			break
		}
		n++
	}
	data := buf.String()
	data = strings.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("empty ssh_info")
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
		sshInfo.ProtoType = config.RemoteTypeGerrit
	} else {
		err = json.Unmarshal([]byte(data), &sshInfo)
		if err != nil {
			return nil, err
		}
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
