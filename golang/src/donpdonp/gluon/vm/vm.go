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

func Factory(name string) (*VM) {
  new_vm := VM{Name: name,
               Js: otto.New()};
  List = append(List, new_vm)
  return &new_vm;
}

func (vm *VM) Load(url string) {
  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("http err")
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  fmt.Println("Otto about to eval:", string(body))
  _, err = vm.Js.Run(body)
  if err != nil {
    fmt.Println("eval failed", err)
  }
}

