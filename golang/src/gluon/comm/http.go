package comm

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func HttpGet(url string, headers map[string]string) (*http.Response, []byte, *tls.ConnectionState, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, nil, err
	} else {
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		return resp, bytes, resp.TLS, err
	}
}

func HttpPost(url string, headers map[string]string, payload io.Reader) (string, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return "", err
	}
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	} else {
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		return string(bytes), err
	}
}
