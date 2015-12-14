package main;

import (
  "fmt"
  "donpdonp/gluon/comm"
  "donpdonp/gluon/vm"
)

func main() {
  bus, _ := comm.Factory()

  go bus.Start("tcp://127.0.0.1:40899")

  fmt.Println("bus started")

  bus.Send(map[string]string{"a":"b"})

  for {
    msg := <-bus.Pipe
    method := msg["method"].(string)
    fmt.Println("method: "+method)

    switch method {
    case "vm.add":
      params := msg["params"].(map[string]interface{})
      url := params["url"].(string)
      vm.Add(vm.VM{}, url)
    case "irc.privmsg":
      for _, vm := range vm.List {
        fmt.Println("VM "+vm.Name)
        vm.Js.Run("1")
      }
    }
  }
}
