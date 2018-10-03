package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"donpdonp/gluon/comm"
	"donpdonp/gluon/util"
	"donpdonp/gluon/vm"

	"github.com/robertkrimen/otto"
)

var (
	vm_list = vm.ListFactory()
	bigben  = make(chan map[string]interface{})
)

func main() {
	util.LoadSettings()

	bus := comm.PubsubFactory(util.Settings.Id)
	busAddr := "localhost:6379"
	fmt.Printf("redis connect %s\n", busAddr)
	bus.Start(busAddr)
	go bus.Loop()

	fmt.Println("gluon started. key " + util.Settings.Key)
	go clocktower(bus)

	for {
		select {
		case msg := <-bus.Pipe:
			if comm.Msg_check(msg) {
				json, _ := json.Marshal(msg)
				fmt.Println("gluon <-", string(json))
				if msg["method"] != nil {
					method := msg["method"].(string)
					rpc_dispatch(bus, method, msg)
				}
			} else {
				fmt.Println(msg)
			}
		case pkt := <-vm_list.Backchan:
			backchan_size := len(vm_list.Backchan)
			if backchan_size > 0 {
  			fmt.Println("backchan queue size ", backchan_size)
  		}
			if pkt["callback"] != nil {
				callback := pkt["callback"].(otto.Value)
  			fmt.Printf("backchan callback pkt: %+v\n", pkt)
				_, err := callback.Call(callback, pkt["result"])
				if err != nil {
					fmt.Println("backchan callback err: " + err.Error())
					//sayback := "["+vm.Name+"] "+err.Error()
					//bus.Send(irc_reply(pkt["msg"], sayback), nil)
				}
			}
		case tick := <-bigben:
			dispatch(tick, bus)
		}
	}

}

func rpc_dispatch(bus comm.Pubsub, method string, msg map[string]interface{}) {
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
	default:
		if key_check(msg) {
			delete(msg, "key")
			dispatch(msg, bus)
		}
	}
}

func key_check(params map[string]interface{}) bool {
	ok := false
	if params["key"] != nil {
		if params["key"] == util.Settings.Key {
			ok = true
		}
	}
	if ok == false {
		fmt.Println("msg.key check failed!")
	}
	return ok
}

func vm_add(owner string, url string, bus comm.Pubsub) error {
	new_vm := vm.Factory(owner)

	vm_enhance_standard(new_vm, bus)

	new_vm.Url = url
	code, err := comm.HttpGet(url)
	if err == nil {
		fmt.Println("vm_add eval")
		err := new_vm.Eval(code)

		if err == nil {
			vm_list.Add(*new_vm)
			fmt.Println("VM " + new_vm.Owner + "/" + new_vm.Name + " added!")
			return nil
		}
		return err
	}
	return err
}

func vm_enhance_standard(vm *vm.VM, bus comm.Pubsub) {
	vm.Js.Set("bot", map[string]interface{}{
		"say": func(call otto.FunctionCall) otto.Value {
			fmt.Printf("gluon say %s %+v\n", call.Argument(0).String(), call.Argument(1).String())
			resp := map[string]interface{}{"method": "irc.privmsg"}
			resp["params"] = map[string]interface{}{"channel": call.Argument(0).String(),
				"message": call.Argument(1).String()}
			bus.Send(resp, nil)
			return otto.Value{}
		},
		"owner":         vm.Owner,
		"admin_channel": util.Settings.AdminChannel,
		"host_id":       util.Settings.Id})
	vm.Js.Set("http", map[string]interface{}{
		"get": func(call otto.FunctionCall) otto.Value {
			fmt.Printf("get(%s)\n", call.Argument(0).String())
			body, err := comm.HttpGet(call.Argument(0).String())
			var ottoStr otto.Value
			if err != nil {
				fmt.Println("http get err")
				ottoStr, _ = otto.ToValue("")
			} else {
				ottoStr, _ = otto.ToValue(body)
			}
			return ottoStr
		},
		"post": func(call otto.FunctionCall) otto.Value {
			fmt.Printf("post(%s, %s)\n", call.Argument(0).String(), call.Argument(1).String())
			body, err := comm.HttpPost(call.Argument(0).String(),
				"application/json",
				strings.NewReader(call.Argument(1).String()))
			var ottoStr otto.Value
			if err != nil {
				fmt.Println("http post err")
				ottoStr, _ = otto.ToValue("")
			} else {
				ottoStr, _ = otto.ToValue(body)
			}
			return ottoStr
		}})
	vm.Js.Set("db", map[string]interface{}{
		"get": func(call otto.FunctionCall) otto.Value {
			resp := map[string]interface{}{"method": "db.get"}
			key := call.Argument(0).String()
			resp["params"] = map[string]interface{}{"group": vm.Owner, "key": key}
			if call.Argument(1).IsDefined() {
				bus.Send(resp, func(pkt map[string]interface{}) {
					pkt["callback"] = call.Argument(1)
					vm_list.Backchan <- pkt
				})
			}
			return otto.Value{}
		},
		"del": func(call otto.FunctionCall) otto.Value {
			resp := map[string]interface{}{"method": "db.del"}
			key := call.Argument(0).String()
			resp["params"] = map[string]interface{}{"group": vm.Owner, "key": key}
			if call.Argument(1).IsDefined() {
				bus.Send(resp, func(pkt map[string]interface{}) {
					pkt["callback"] = call.Argument(1)
					vm_list.Backchan <- pkt
				})
			}
			return otto.Value{}
		},
		"len": func(call otto.FunctionCall) otto.Value {
			resp := map[string]interface{}{"method": "db.len"}
			resp["params"] = map[string]interface{}{"group": vm.Owner}
			if call.Argument(0).IsDefined() {
				bus.Send(resp, func(pkt map[string]interface{}) {
					pkt["callback"] = call.Argument(0)
					vm_list.Backchan <- pkt
				})
			}
			return otto.Value{}
		},
		"set": func(call otto.FunctionCall) otto.Value {
			key := call.Argument(0).String()
			value := call.Argument(1).String()
			resp := map[string]interface{}{"method": "db.set"}
			resp["params"] = map[string]interface{}{"group": vm.Owner, "key": key, "value": value}
			bus.Send(resp, func(pkt map[string]interface{}) {
				if call.Argument(2).IsDefined() {
					pkt["callback"] = call.Argument(2)
					vm_list.Backchan <- pkt
				}
			})
			return otto.Value{}
		},
		"scan": func(call otto.FunctionCall) otto.Value {
			cursor := call.Argument(0).String()
			match := call.Argument(1).String()
			count := call.Argument(2).String()
			resp := map[string]interface{}{"method": "db.scan"}
			resp["params"] = map[string]interface{}{"group": vm.Owner, "cursor": cursor, "match": match, "count": count}
			bus.Send(resp, func(pkt map[string]interface{}) {
				if call.Argument(3).IsDefined() {
					pkt["callback"] = call.Argument(3)
					vm_list.Backchan <- pkt
				}
			})
			return otto.Value{}
		}})
	vm.Js.Set("vm", map[string]interface{}{
		"add": func(call otto.FunctionCall) otto.Value {
			url := call.Argument(0).String()
			err := vm_add(vm.Owner, url, bus)
			if err == nil {
				return otto.Value{}
			} else {
				otto_str, _ := otto.ToValue(err.Error())
				return otto_str
			}
		},
		"del": func(call otto.FunctionCall) otto.Value {
			name := call.Argument(0).String()
			url, err := vm_del(name, bus)
			if err == nil {
				otto_str, _ := otto.ToValue(url)
				return otto_str
			} else {
				err_str, _ := otto.ToValue(err.Error())
				return err_str
			}
		},
		"reload": func(call otto.FunctionCall) otto.Value {
			name := call.Argument(0).String()
			err := vm_reload(name, bus)
			if err == nil {
				return otto.Value{}
			} else {
				otto_str, _ := otto.ToValue(err.Error())
				return otto_str
			}
		},
		"list": func(call otto.FunctionCall) otto.Value {
			vm_names := []string{}
			for vm := range vm_list.Range() {
				vm_names = append(vm_names, vm.Owner+"/"+vm.Name)
			}
			callback := call.Argument(0)
			callback.Call(callback, vm_names)
			return otto.Value{}
		}})
}

func vm_reload(name string, bus comm.Pubsub) error {
	idx := vm_list.IndexOf(name)
	if idx > -1 {
		vm := vm_list.At(idx)
		fmt.Println(name + " found. reloading " + vm.Url)
		code, err := comm.HttpGet(vm.Url)
		if err != nil {
			err := vm.Eval(code)
			return err
		}
	}
	return errors.New(name + " not found")
}

func vm_del(name string, bus comm.Pubsub) (string, error) {
	idx := vm_list.IndexOf(name)
	if idx > -1 {
		url, err := vm_list.Del(name)
		if err == nil {
			fmt.Println(name + " deleted.")
			return url, nil
		}
		return "", err
	}
	fmt.Println(name + " not found.")
	return "", errors.New("not found")
}

func do_vm_list(bus comm.Pubsub) {
	fmt.Println("VM List")
	for vm := range vm_list.Range() {
		fmt.Println("* " + vm.Owner + "/" + vm.Name)
	}
	fmt.Println("VM List done")
}

func dispatch(msg map[string]interface{}, bus comm.Pubsub) {
	for vm := range vm_list.Range() {
		pprm, _ := json.Marshal(msg)
		call_js := "go(" + string(pprm) + ")"
		fmt.Printf("** %s/%s %s\n", vm.Owner, vm.Name, call_js)
		value, err := vm.Js.Run(call_js)
		if msg["method"] == "irc.privmsg" {
			var sayback string
			if err != nil {
				sayback = "[" + vm.Name + "] " + err.Error()
			} else {
				if value.IsDefined() {
					sayback = value.String()
				}
			}
			if len(sayback) > 0 {
				bus.Send(irc_reply(msg, sayback), nil)
			}
		}
	}
}

func irc_reply(msg map[string]interface{}, value string) map[string]interface{} {
	params := msg["params"].(map[string]interface{})
	resp := map[string]interface{}{"method": "irc.privmsg"}

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
		msg := map[string]interface{}{"method": "clocktower"}
		msg["params"] = map[string]interface{}{"time": time.Now().UTC().Format("2006-01-02T15:04:05Z")}
		bigben <- msg

		time.Sleep(60 * time.Second)
	}
}
