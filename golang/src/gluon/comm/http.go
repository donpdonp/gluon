package comm

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func HttpGet(url string) (*http.Response, []byte, *tls.ConnectionState, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, nil, nil, err
	} else {
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		return resp, bytes, resp.TLS, err
	}
}

func HttpPost(url string, mime string, payload io.Reader) (string, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Post(url, mime, payload)
	if err != nil {
		return "", err
	} else {
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		return string(bytes), err
	}
}
