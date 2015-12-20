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
    if comm.Msg_check(msg, my_uuid.String()) {
      json, _ := json.Marshal(msg)
      fmt.Println("<-", string(json))
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

func vm_add(name string, url string, bus comm.Pubsub, my_uuid string) {
  new_vm := vm.Factory(name)
  new_vm.Js.Set("say", func(call otto.FunctionCall) otto.Value {
      fmt.Printf("say %s %s.\n", call.Argument(0).String(), call.Argument(1).String(), call.Argument(2).String())
      resp := map[string]interface{}{"id": uuid.NewV4(), "from": my_uuid, "method":"irc.privmsg"}
      resp["params"] = map[string]interface{}{"irc_session_id":call.Argument(0).String(),
                                              "channel":call.Argument(1).String(),
                                              "message": call.Argument(2).String()}
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