package main;

import (
  "fmt"
  "encoding/json"

  "donpdonp/gluon/comm"
  "donpdonp/gluon/vm"

  "github.com/robertkrimen/otto"
  "github.com/satori/go.uuid"
)

func main() {
  bus := comm.PubsubFactory()
  my_uuid := uuid.NewV4()
  bus.Start("localhost:6379")
  go bus.Loop()

  fmt.Println("bus started")

  for {
    msg := <-bus.Pipe
    if msg_check(msg, my_uuid.String()) {
      //id := msg["id"].(string)
      method := msg["method"].(string)
      fmt.Println("method: "+method)

      switch method {
      case "vm.add":
        params := msg["params"].(map[string]interface{})
        url := params["url"].(string)
        name := params["name"].(string)
        vm_add(name, url, bus, my_uuid.String())
      case "irc.privmsg":
        dispatch(msg, bus, my_uuid.String())
      }
    }
  }

}

func msg_check(msg map[string]interface{}, my_uuid string) (bool) {
  if msg["id"] != nil && msg["from"] != nil {
    if msg["method"] != nil || msg["result"] != nil || msg["error"] != nil {
      from := msg["from"].(string)
      if from == my_uuid {
        // drop my own msgs
        fmt.Println("dropping my own echo")
        return false
      } else {
        return true
      }
    } else {
      fmt.Println("missing msg method/result/error!")
      return false
    }
  } else {
    fmt.Println("missing msg id/from!")
    return false
  }
}

func vm_add(name string, url string, bus comm.Pubsub, my_uuid string) {
  new_vm := vm.Factory(name)
  new_vm.Js.Set("say", func(call otto.FunctionCall) otto.Value {
      fmt.Printf("Hello, %s.\n", call.Argument(0).String())
      resp := map[string]interface{}{"id": uuid.NewV4(), "from": my_uuid, "method":"irc.privmsg"}
      resp["params"] = map[string]interface{}{"message": call.Argument(0).String()}
      bus.Send(resp)
      return otto.Value{}
  })
  new_vm.Load(url)
}

func dispatch(msg map[string]interface{}, bus comm.Pubsub, my_uuid string) {
  for _, vm := range vm.List {
    fmt.Println("VM "+vm.Name)
    pprm, _ := json.Marshal(msg)
    call_js := "go("+string(pprm)+")"
    fmt.Println("js call: "+call_js)
    value, err := vm.Js.Run(call_js)
    if err != nil {
      bus.Send(map[string]interface{}{"id": uuid.NewV4(), "from": my_uuid, "error":err.Error()})
    } else {
      bus.Send(map[string]interface{}{"id": uuid.NewV4(), "from": my_uuid, "result":value.String()})
    }
  }
}