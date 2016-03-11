package comm

import (
  "net/http"
  "io/ioutil"
)


func HttpGet(url string) (string, error) {
  resp, err := http.Get(url)
  if err != nil {
    return "", err
  } else {
    defer resp.Body.Close()
    bytes, err := ioutil.ReadAll(resp.Body)
    return string(bytes), err
  }
}
