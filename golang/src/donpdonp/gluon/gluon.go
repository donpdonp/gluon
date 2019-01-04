package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
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
			pipe_size := len(bus.Pipe)
			if pipe_size > 0 {
				fmt.Printf("msg pipe backlog: %d\n", pipe_size)
			}
			if comm.Msg_check(msg) {
				json, _ := json.Marshal(msg)
				fmt.Println("gluon <-", string(json))
				if msg["method"] != nil {
					method := msg["method"].(string)
					rpc_dispatch(bus, method, msg)
				}
			} else {
				fmt.Printf("msg not understood: %#v\n", msg)
			}
		case pkt := <-vm_list.Backchan:
			backchan_size := len(vm_list.Backchan)
			if backchan_size > 0 {
				fmt.Println("VM callback queue backlog ", backchan_size)
			}
			if pkt["callback"] != nil {
				callback := pkt["callback"].(otto.Value) //otto.FunctionCall
				_, err := callback.Call(callback, pkt["result"])
				if err != nil {
					vm_name := pkt["vm"].(string)
					sayback := err.Error()
					fmt.Printf("backchan callback err: %s %#v\n", vm_name, err)
					fakemsg := map[string]interface{}{"params": map[string]interface{}{
						                "channel": util.Settings.AdminChannel}}
					bus.Send(irc_reply(fakemsg, vm_name+" "+sayback, vm_name), nil)
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

func make_callback(pkt map[string]interface{}, cb otto.Value, vm *vm.VM) {
	pkt["callback"] = cb
	pkt["vm"] = vm.Owner + "/" + vm.Name
}

func vm_enhance_js_standard(vm *vm.VM, bus comm.Pubsub) {
	vm.Js.Set("bot", map[string]interface{}{
		"say": func(call otto.FunctionCall) otto.Value {
			fmt.Printf("irc.privmsg %s %+v\n", call.Argument(0).String(), call.Argument(1).String())
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
			fmt.Printf("http.get %s\n", call.Argument(0).String())
			_, body, err := comm.HttpGet(call.Argument(0).String())
			var ottoStr otto.Value
			if err != nil {
				fmt.Printf("http.get err %v\n", err)
				ottoStr, _ = otto.ToValue("")
			} else {
				fmt.Printf("http.get OK %d bytes\n", len(body))
				ottoStr, _ = otto.ToValue(string(body)) // scripts not ready for bytes
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
			fmt.Printf("db.get %s\n", call.Argument(0).String())
			resp := map[string]interface{}{"method": "db.get"}
			key := call.Argument(0).String()
			resp["params"] = map[string]interface{}{"group": vm.Owner, "key": key}
			if call.Argument(1).IsDefined() {
				bus.Send(resp, func(pkt map[string]interface{}) {
					make_callback(pkt, call.Argument(1), vm)
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
					make_callback(pkt, call.Argument(1), vm)
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
					make_callback(pkt, call.Argument(0), vm)
					vm_list.Backchan <- pkt
				})
			}
			return otto.Value{}
		},
		"set": func(call otto.FunctionCall) otto.Value {
			fmt.Printf("db.set %s\n", call.Argument(0).String())
			key := call.Argument(0).String()
			value := call.Argument(1).String()
			resp := map[string]interface{}{"method": "db.set"}
			resp["params"] = map[string]interface{}{"group": vm.Owner, "key": key, "value": value}
			bus.Send(resp, func(pkt map[string]interface{}) {
				if call.Argument(2).IsDefined() {
					make_callback(pkt, call.Argument(2), vm)
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
					make_callback(pkt, call.Argument(3), vm)
					vm_list.Backchan <- pkt
				}
			})
			return otto.Value{}
		}})
	vm.Js.Set("vm", map[string]interface{}{
		"add": func(call otto.FunctionCall) otto.Value {
			url := call.Argument(0).String()
			descriptor, err := vm_add(vm.Owner, url, bus)
			if err == nil {
				ottoDescriptor, _ := otto.ToValue(descriptor)
				return ottoDescriptor
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
			_, err := callback.Call(callback, vm_names)
			if err != nil {
				fmt.Printf("vm list callback err: %s\n", err.Error())
				otto_str, _ := otto.ToValue(err.Error())
				return otto_str
			} else {
				return otto.Value{}
			}
		}})
}

func vm_enhance_ruby_standard(vmm *vm.VM, bus comm.Pubsub) {
	vm.RubyStdCallbacks(vmm, func(channel string, say string) {
		fmt.Printf("gluon ruby say %#v %#v\n", channel, say)
		resp := map[string]interface{}{"method": "irc.privmsg"}
		resp["params"] = map[string]interface{}{"channel": channel,
			"message": say}
		bus.Send(resp, nil)
	})
}

func vm_add(owner string, url string, bus comm.Pubsub) (map[string]interface{}, error) {
	fmt.Printf("--vm_add owner: %v url: %v\n", owner, url)
	resp, codeBytes, err := comm.HttpGet(url)
	if err != nil {
		fmt.Printf("vm_add http err %v\n", err)
	} else {
		len := 0
		if resp.Header["Content-Length"] != nil {
			len, _ = strconv.Atoi(resp.Header["Content-Length"][0])
		}
		lang := "undefined"
		if resp.Header["Content-Type"] != nil {
			lang = pickLang(url, resp.Header["Content-Type"][0])
		}
		fmt.Printf("vm_add %s http %s %d bytes\n", lang, resp.Header["Content-Type"], len)
		vm := vm.Factory(owner, lang)
		vm.Url = url
		var setup_json string
		if vm.Lang() == "javascript" {
			vm_enhance_js_standard(vm, bus)
			setup_json, err = vm.FirstEvalJs(string(codeBytes))
		} else if vm.Lang() == "ruby" {
			//vm_enhance_ruby_standard(vm, bus)
			//setup_json, err = vm.Eval(code)
			err = errors.New("no ruby support.")
		} else if vm.Lang() == "webassembly" {
			setup_json, _ = vm.Eval(codeBytes)
		} else {
			err = errors.New("unknown lang " + lang)
		}
		if err != nil {
			fmt.Printf("vm_add err: %v\n", err)
		} else {
			var setup map[string]interface{}
			json.Unmarshal([]byte(setup_json), &setup)
			json_str, _ := json.Marshal(setup)
			fmt.Printf("setup_json %s\n", json_str)
			vm.Name = setup["name"].(string)
			vm_list.Add(*vm)
			fmt.Printf("VM %s/%s (%s) added!\n", vm.Owner, vm.Name, vm.Lang())
			return setup, nil
		}
		return nil, err
	}
	return nil, err
}

func pickLang(urlStr string, contentType string) string {
	var lang string
	if contentType == "script/javascript" {
		lang = "javascript"
	} else if contentType == "script/ruby" {
		lang = "ruby"
	} else {
		uri, _ := url.Parse(urlStr)
		parts := strings.Split(uri.Path, "/")
		filename := parts[len(parts)-1]
		filename_parts := strings.Split(filename, ".")
		extension := filename_parts[len(filename_parts)-1]
		if extension == "js" {
			lang = "javascript"
		} else if extension == "rb" {
			lang = "ruby"
		} else if extension == "wasm" {
			lang = "webassembly"
		} else if extension == "wast" { //webasm source
			lang = "webassembly"
		}
	}
	return lang
}

func vm_reload(name string, bus comm.Pubsub) error {
	idx := vm_list.IndexOf(name)
	if idx > -1 {
		vm := vm_list.At(idx)
		fmt.Println(name + " found. reloading " + vm.Url)
		_, codeBytes, err := comm.HttpGet(vm.Url)
		if err != nil {
			_, err := vm.Eval(codeBytes)
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
	fmt.Printf("[* dispatch %s to %d VMs\n", msg["method"], vm_list.Size())
	for vm := range vm_list.Range() {
		start := time.Now()
		params_jbytes, _ := json.Marshal(msg)
		params_json := string(params_jbytes)
		var callBytes []byte
		if vm.Lang() == "javascript" {
			callBytes = []byte("go(" + params_json + ")")
		}
		if vm.Lang() == "ruby" {
			params_double_jbytes, _ := json.Marshal(params_json)
			params_double_json := string(params_double_jbytes)
			callBytes = []byte("go(JSON.parse(" + params_double_json + "))")
		}
		json_str, err := vm.Eval(callBytes)
		elapsed := time.Now().Sub(start)
		var sayback string
		if err != nil {
			fmt.Printf("** %s/%s dispatch err: %v\n", vm.Owner, vm.Name, err)
			sayback = "[" + vm.Name + "] " + err.Error()
		} else {
			//fmt.Printf("** %s/%s dispatch call return json: %#v\n", vm.Owner, vm.Name, json_str)
			var said interface{}
			err := json.Unmarshal([]byte(json_str), &said)
			fmt.Printf("** %s/%s %#v [%.4f sec]\n", vm.Owner, vm.Name, said, elapsed.Seconds())
			if err == nil {
				if said != nil {
					sayback = said.(string)
				}
			}
		}
		if len(sayback) > 0 {
			if msg["method"] != "irc.privmsg" {
				if msg["params"] == nil {
					msg["params"] = map[string]interface{}{}
				}
				msg["params"].(map[string]interface{})["channel"] = util.Settings.AdminChannel
			}
			bus.Send(irc_reply(msg, sayback, vm.Owner+"/"+vm.Name), nil)
		}
	}
}

func irc_reply(msg map[string]interface{}, value string, vm_name string) map[string]interface{} {
	params := msg["params"].(map[string]interface{})
	resp := map[string]interface{}{"method": "irc.privmsg"}

	out := params["channel"].(string)
	if out[0:1] != "#" {
		out = params["nick"].(string)
	}

	fmt.Printf("%s %s irc.privmsg: %s\n", vm_name, out, value)
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
