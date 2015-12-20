package main;

import (
  "fmt"
  "encoding/json"

  "donpdonp/gluon/comm"
  "donpdonp/gluon/vm"
)

func main() {
  bus := comm.PubsubFactory()

  bus.Start("localhost:6379")
  go bus.Loop()

  fmt.Println("bus started")

  for {
    msg := <-bus.Pipe
    mo := msg["method"]
    if mo != nil {
      method := mo.(string)
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
          pprm, _ := json.Marshal(msg)
          call_js := "go("+string(pprm)+")"
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

}
