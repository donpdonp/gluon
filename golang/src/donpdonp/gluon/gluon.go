package main;

import (
  "fmt"
  "time"
  "encoding/json"
  "net/http"
  "io/ioutil"

  "donpdonp/gluon/comm"
  "donpdonp/gluon/vm"

  "github.com/robertkrimen/otto"
  "github.com/satori/go.uuid"
)

var (
  my_uuid uuid.UUID = uuid.NewV4()
)

func main() {
  bus := comm.PubsubFactory(my_uuid.String())
  bus.Start("localhost:6379")
  go bus.Loop()

  fmt.Println("gluon started")
  go clocktower(bus)

  for {
    msg := <-bus.Pipe
    if comm.Msg_check(msg) {
      json, _ := json.Marshal(msg)
      fmt.Println("<-", string(json))
      if msg["method"] != nil {
        method := msg["method"].(string)
        fmt.Println("method: "+method)

        switch method {
        case "vm.add":
          params := msg["params"].(map[string]interface{})
          url := params["url"].(string)
          owner := params["owner"].(string)
          vm_add(owner, url, bus, my_uuid.String())
        case "irc.privmsg":
          dispatch(msg, bus)
        }
      }
    } else {
      fmt.Println(msg)
    }
  }

}

func vm_add(owner string, url string, bus comm.Pubsub, my_uuid string) {
  new_vm := vm.Factory(owner)
  new_vm.Js.Set("bot", map[string]interface{}{"say":func(call otto.FunctionCall) otto.Value {
      fmt.Printf("say(%s %s %s)\n", call.Argument(0).String(), call.Argument(1).String(), call.Argument(2).String())
      resp := map[string]interface{}{"method":"irc.privmsg"}
      resp["params"] = map[string]interface{}{"channel":call.Argument(0).String(),
                                              "message": call.Argument(1).String()}
      bus.Send(resp, nil)
      return otto.Value{}
  }})
  new_vm.Js.Set("http", map[string]interface{}{"get":func(call otto.FunctionCall) otto.Value {
      fmt.Printf("get(%s)\n", call.Argument(0).String())
      resp, err := http.Get(call.Argument(0).String())
      if err != nil {
        fmt.Println("http err")
      }
      defer resp.Body.Close()
      body, err := ioutil.ReadAll(resp.Body)
      ottoStr, _ := otto.ToValue(string(body))
      return ottoStr
  }})
  new_vm.Js.Set("db", map[string]interface{}{
    "get":func(call otto.FunctionCall) otto.Value {
      fmt.Printf("get(%s)\n", call.Argument(0).String())
      resp := map[string]interface{}{"method":"db.get"}
      key := call.Argument(0).String()
      resp["params"] = map[string]interface{}{"group":new_vm.Owner, "key":key}
      bus.Send(resp, func(pkt map[string]interface{}){
        fmt.Println(pkt["result"])
      })
      return otto.Value{}
    },
    "set":func(call otto.FunctionCall) otto.Value {
      key := call.Argument(0).String()
      value := call.Argument(1).String()
      fmt.Printf("set(%s, %s)\n", key, value)
      resp := map[string]interface{}{"method":"db.set"}
      resp["params"] = map[string]interface{}{"group":new_vm.Owner, "key":key, "value": value}
      bus.Send(resp, func(pkt map[string]interface{}){
        fmt.Println(pkt["result"])
      })
      return otto.Value{}
    }})
  new_vm.Load(url)
  vm.List = append(vm.List, *new_vm)
}

func dispatch(msg map[string]interface{}, bus comm.Pubsub) {
  for _, vm := range vm.List {
    pprm, _ := json.Marshal(msg)
    call_js := "go("+string(pprm)+")"
    fmt.Println("**VM", vm.Owner, "/", vm.Name, ": ", call_js)
    value, err := vm.Js.Run(call_js)
    if err != nil {
      bus.Send(irc_reply(msg, err.Error()), nil)
    } else {
      if value.IsDefined() {
        bus.Send(irc_reply(msg, value.String()), nil)
      }
    }
  }
}

func irc_reply(msg map[string]interface{}, value string) (map[string]interface{}) {
  params := msg["params"].(map[string]interface{})
  resp := map[string]interface{}{"method":"irc.privmsg"}

  resp["params"] = map[string]interface{}{"irc_session_id": params["irc_session_id"],
                                          "channel": params["channel"],
                                          "message": value}
  return resp
}

func clocktower(bus comm.Pubsub) {
  fmt.Println("clocktower started", time.Now())
  for {
    msg := map[string]interface{}{"method":"clocktower"}
    msg["params"] = map[string]interface{}{"time": time.Now().UTC().Format("2006-01-02T15:04:05Z")}
    dispatch(msg, bus)

    //bus.Send(msg)
    time.Sleep(60 * time.Second)
  }
}
