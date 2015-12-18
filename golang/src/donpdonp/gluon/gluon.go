package main;

import (
  "fmt"
  "donpdonp/gluon/comm"
  "donpdonp/gluon/vm"
)

func main() {
  bus, _ := comm.Factory()

  bus.Start("tcp://127.0.0.1:40899")
  go bus.Loop()

  fmt.Println("bus started")

  for {
    msg := <-bus.Pipe
    method := msg["method"].(string)
    fmt.Println("method: "+method)

    switch method {
    case "vm.add":
      params := msg["params"].(map[string]interface{})
      url := params["url"].(string)
      name := params["name"].(string)
      new_vm := vm.Factory(name)
      new_vm.Load(url)
    case "irc.privmsg":
      for _, vm := range vm.List {
        fmt.Println("VM "+vm.Name)
        params := msg["params"].(map[string]interface{})
        p1 := params["msg"].(string)
        call_js := "go(\""+p1+"\")"
        fmt.Println("js call: "+call_js)
        value, err := vm.Js.Run(call_js)
        if err != nil {
          bus.Send(map[string]interface{}{"error":err.Error()})
        } else {
          bus.Send(map[string]interface{}{"result":value.String()})
        }
      }
    }
  }

}
