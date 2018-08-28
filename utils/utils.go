package utils

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetNextProxyHeader will parse the request header and send request
// to next agent based header infomation
func GetNextProxyHeader(r *http.Request) ([]string, error) {
	currentProxyHeader := r.Header.Get("X-Prometheus-Proxy")
	currentRequestHeader := r.Header.Get("X-Prometheus-Request")
	if currentRequestHeader == "" {
		return nil, errors.New("Proxy Chain Broken")
	}
	if currentProxyHeader == "" {
		return []string{"", "", currentRequestHeader}, nil
	}
	proxyList := strings.Split(currentProxyHeader, ";")
	nextStop := proxyList[0]
	nextProxyHeader := ""
	if len(proxyList) > 1 {
		nextProxyHeader = strings.Join(proxyList[1:], ";")
	}
	return []string{nextStop, nextProxyHeader, currentRequestHeader}, nil
}

// MakeRequest make http request with headers
func MakeRequest(url string, header map[string]string) (string, error) {
	log.Infof("Send proxy request to %s", url)
	log.Infof("Request Header is %v", header)
	if url == "" {
		return "", errors.New("url is empty")
	}

	if !strings.HasPrefix(url, "http://") {
		url = "http://" + url
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// MakePrometheusRequest send http request
func MakePrometheusRequest(r *http.Request) (string, error) {
	parseResult, err := GetNextProxyHeader(r)
	if err != nil {
		return "", err
	}
	// Direct send request to local prometheus server
	if parseResult[0] == "" && parseResult[1] == "" && parseResult[2] != "" {
		resp, err := MakeRequest(parseResult[2], nil)
		if err != nil {
			return "", err
		}
		return resp, nil
	}
	if parseResult[0] != "" {
		resp, err := MakeRequest(parseResult[0], map[string]string{
			"X-Prometheus-Proxy":   parseResult[1],
			"X-Prometheus-Request": parseResult[2],
		})
		if err != nil {
			return "", err
		}
		return resp, nil
	}
	return "", errors.New("Http headers unknown")
}
