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

func Factory(owner string) (*VM) {
  new_vm := VM{Owner: owner,
               Js: otto.New()};
  return &new_vm;
}

func (vm *VM) Load(url string) bool {
  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("http err")
  } else {
    defer resp.Body.Close()
    vm.Url = url
    body, err := ioutil.ReadAll(resp.Body)
    fmt.Println(string(body))
    fmt.Println("--eval begins--")

    src, err := vm.Js.Compile("", body)

    if err != nil {
      fmt.Println("compile failed!", err)
    } else {
      fmt.Println("compile good!")
      setup, err := vm.Js.Run(src)
      if err != nil {
        fmt.Println("eval failed", err, vm.Js.Context().Stacktrace)
      } else {
        descriptor_value, err := setup.Call(setup)
        if err != nil {
          fmt.Println("js func setup eval fail")
        } else {
          descriptor_map, _ := descriptor_value.Export()
          descriptor := descriptor_map.(map[string]interface{})
          vm.Name = descriptor["name"].(string)
          return true
        }
      }
    }
  }
  return false
}

