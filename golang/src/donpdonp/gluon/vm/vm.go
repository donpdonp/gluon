package vm

import (
  "fmt"
  "net/http"
  "io/ioutil"

  "github.com/robertkrimen/otto"
)

type VM struct {
  Owner string
  Name string
  Url string
  Js *otto.Otto
}

var (
  List []VM
)

func Factory(owner string) (*VM) {
  new_vm := VM{Owner: owner,
               Js: otto.New()};
  return &new_vm;
}

func (vm *VM) Load(url string) {
  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("http err")
  } else {
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    fmt.Println("Otto about to eval:", string(body))

    descriptor_value, err := vm.Js.Run(body)
    if err != nil {
      fmt.Println("eval failed", err)
    } else {
      descriptor_map, _ := descriptor_value.Export()
      descriptor := descriptor_map.(map[string]interface{})
      vm.Name = descriptor["name"].(string)
    }
  }
}
