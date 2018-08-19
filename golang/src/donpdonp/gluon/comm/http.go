package comm

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func HttpGet(url string) (string, error) {
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	} else {
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		return string(bytes), err
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
