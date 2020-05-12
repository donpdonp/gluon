package main

import "fmt"
import "strings"
import "time"

import "donpdonp/gluon/comm"
import "donpdonp/gluon/vm"
import "github.com/robertkrimen/otto"

func httpGet(vm *vm.VM, call otto.FunctionCall) otto.Value {
	var ottoV otto.Value
	arg0 := call.Argument(0)
	if arg0.IsString() {
		urlc := arg0.String()
		resp, body, _, err := comm.HttpGet(urlc)
		var resultDisplay string
		if err != nil {
			resultDisplay = fmt.Sprintf("err %#v\n", err)
			ottoV, _ = otto.ToValue("")
		} else {
			ottoV, _ = otto.ToValue(string(body)) // scripts not ready for bytes
			resultDisplay = fmt.Sprintf("%s %d bytes\n", resp.Status, len(body))
		}
		fmt.Printf("%s/%s http.get %s %s\n", vm.Owner, vm.Name, urlc, resultDisplay)
	} else if arg0.IsObject() {
		urlo := arg0.Object()
		urlg, _ := urlo.Get("url")
		urlc := urlg.String()
		fmt.Printf("%s/%s http.get %#v\n", vm.Owner, vm.Name, urlc)
		resp, body, tls, err := comm.HttpGet(urlc)
		fmt.Printf("go %#v %#v\n", resp, err)
		goResult := map[string]interface{}{}
		if err != nil {
			goResult["error"] = map[string]interface{}{
				"message": err.Error()}
		} else {
			goResult["status"] = resp.StatusCode
			goResult["body"] = string(body)
			goTls := map[string]interface{}{}
			if tls != nil {
				goTls["version"] = tls.Version
				goTls["server_name"] = tls.ServerName
				certs := []map[string]interface{}{}
				for _, cert := range tls.PeerCertificates {
					c := map[string]interface{}{}
					c["not_before"] = cert.NotBefore.Format(time.RFC3339)
					c["not_after"] = cert.NotAfter.Format(time.RFC3339)
					c["dns_names"] = cert.DNSNames
					certs = append(certs, c)
				}
				goTls["peer_certs"] = certs
			}
			goResult["tls"] = goTls
		}
		ottoV, err = vm.Js.ToValue(goResult)
		fmt.Printf("otto %#v %#v\n", ottoV, err)
	}
	return ottoV
}

func httpPost(vm *vm.VM, call otto.FunctionCall) otto.Value {
	urlc := call.Argument(0).String()
	fmt.Printf("post(%s, %s)\n", urlc, call.Argument(1).String())
	body, err := comm.HttpPost(urlc,
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
}
