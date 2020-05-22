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
		headers := map[string]string{}
		resp, body, _, err := comm.HttpGet(urlc, headers)
		var resultDisplay string
		if err != nil {
			resultDisplay = fmt.Sprintf("err %#v\n", err)
			ottoV, _ = otto.ToValue("") // scripts use length 0 as err indication
		} else {
			ottoV, _ = otto.ToValue(string(body))
			resultDisplay = fmt.Sprintf("%s %d bytes\n", resp.Status, len(body))
		}
		fmt.Printf("%s/%s http.get %s %s\n", vm.Owner, vm.Name, urlc, resultDisplay)
	} else if arg0.IsObject() {
		urlo := arg0.Object()
		urlg, _ := urlo.Get("url")
		urlc := urlg.String()
		headerv, _ := urlo.Get("headers")
		headero := headerv.Object()
		headers := map[string]string{}
		for _, key := range headero.Keys() {
			value, _ := headero.Get(key)
			headers[key] = value.String()
		}
		fmt.Printf("%s/%s http.get %#v\n", vm.Owner, vm.Name, urlc)
		resp, body, tls, err := comm.HttpGet(urlc, headers)
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
	headers := map[string]string{}
	body, err := comm.HttpPost(urlc, headers,
		strings.NewReader(call.Argument(1).String()))
	var resultDisplay string
	var ottoStr otto.Value
	if err != nil {
		resultDisplay = fmt.Sprintf("err %#v\n", err)
		ottoStr, _ = otto.ToValue("")
	} else {
		ottoStr, _ = otto.ToValue(body)
	}
	fmt.Printf("%s/%s http.post %s %s\n", vm.Owner, vm.Name, urlc, resultDisplay)
	return ottoStr
}
