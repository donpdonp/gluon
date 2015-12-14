package vm

import (
  "fmt"
  "net/http"
  "io/ioutil"

  "github.com/robertkrimen/otto"
)

type VM struct {
  Name string
  Url string
  Js *otto.Otto
}

var (
  List []VM
)

func Add(newvm VM, url string) {
  fmt.Println("adding "+url )
  List = append(List, newvm)

  newvm.Name = "Name"
  newvm.Js = otto.New()
  newvm.Url = url
  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("http err")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  newvm.Js.Run(body)
}
