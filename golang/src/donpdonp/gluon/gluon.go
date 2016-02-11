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
  vm_list vm.List
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
          vm_add(owner, url, bus)
        case "vm.reload":
          params := msg["params"].(map[string]interface{})
          name := params["name"].(string)
          vm_reload(name, bus)
        case "vm.del":
          params := msg["params"].(map[string]interface{})
          name := params["name"].(string)
          vm_del(name, bus)
        case "vm.list":
          do_vm_list(bus)
        case "irc.privmsg":
          dispatch(msg, bus)
        }
      }
    } else {
      fmt.Println(msg)
    }
  }

}

func vm_add(owner string, url string, bus comm.Pubsub) bool {
  new_vm := vm.Factory(owner)
  new_vm.Js.Set("bot", map[string]interface{}{"say":func(call otto.FunctionCall) otto.Value {
      fmt.Printf("say(%s %s %s)\n", call.Argument(0).String(), call.Argument(1).String(), call.Argument(2).String())
      resp := map[string]interface{}{"method":"irc.privmsg"}
      resp["params"] = map[string]interface{}{"channel":call.Argument(0).String(),
                                              "message": call.Argument(1).String()}
      bus.Send(resp, nil)
      return otto.Value{}
  }, "owner": new_vm.Owner})
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
      resp := map[string]interface{}{"method":"db.get"}
      key := call.Argument(0).String()
      resp["params"] = map[string]interface{}{"group":new_vm.Owner, "key":key}
      bus.Send(resp, func(pkt map[string]interface{}){
        callback := call.Argument(1)
        callback.Call(callback, pkt["result"])
      })
      return otto.Value{}
    },
    "set":func(call otto.FunctionCall) otto.Value {
      key := call.Argument(0).String()
      value := call.Argument(1).String()
      resp := map[string]interface{}{"method":"db.set"}
      resp["params"] = map[string]interface{}{"group":new_vm.Owner, "key":key, "value": value}
      bus.Send(resp, func(pkt map[string]interface{}){
      })
      return otto.Value{}
    }})
  eval := new_vm.Load(url)
  if eval {
    success, _ := vm_list.Add(*new_vm)
    fmt.Println("VM "+new_vm.Owner+"/"+new_vm.Name+" added!")
    return success
  }
  return false
}

func vm_reload(name string, bus comm.Pubsub) bool {
  idx := vm_list.IndexOf(name)
  if idx >-1 {
    vm := vm_list.At(idx)
    fmt.Println(name+" found. reloading "+vm.Url)
    vm.Load(vm.Url)
    return true
  }
  fmt.Println(name+" not found.")
  return false
}

func vm_del(name string, bus comm.Pubsub) bool {
  idx := vm_list.IndexOf(name)
  if idx >-1 {
    vm_list.Del(name)
    fmt.Println(name+" deleted.")
    return true
  }
  fmt.Println(name+" not found.")
  return false
}

func do_vm_list(bus comm.Pubsub) {
  fmt.Println("VM List")
  for vm := range vm_list.Range() {
    fmt.Println("* "+vm.Owner+"/"+vm.Name)
  }
  fmt.Println("VM List done")
}

func dispatch(msg map[string]interface{}, bus comm.Pubsub) {
  for vm := range vm_list.Range() {
    pprm, _ := json.Marshal(msg)
    call_js := "go("+string(pprm)+")"
    fmt.Println("**VM", vm.Owner, "/", vm.Name, ": ", call_js)
    value, err := vm.Js.Run(call_js)
    if msg["method"] == "irc.privmsg" {
      var sayback string
      if err != nil {
        sayback = "["+vm.Name+"] "+err.Error()
      } else {
        if value.IsDefined() {
          sayback = value.String()
        }
      }
      bus.Send(irc_reply(msg, sayback), nil)
    }
  }
}

func irc_reply(msg map[string]interface{}, value string) (map[string]interface{}) {
  params := msg["params"].(map[string]interface{})
  resp := map[string]interface{}{"method":"irc.privmsg"}

  out := params["channel"].(string)
  if out[0:1] != "#" {
    out = params["nick"].(string)
  }

  resp["params"] = map[string]interface{}{"irc_session_id": params["irc_session_id"],
                                          "channel": out,
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
