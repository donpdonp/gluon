package main

import "fmt"
import "strings"
import "time"
import "crypto/tls"
import "encoding/json"

import "donpdonp/gluon/comm"
import "donpdonp/gluon/vm"
import "github.com/robertkrimen/otto"

func paramParse(arg0 otto.Value) (string, map[string]string) {
	url, headers := "", map[string]string{}
	if arg0.IsString() {
		url = arg0.String()
	} else if arg0.IsObject() {
		urlo := arg0.Object()
		urlg, _ := urlo.Get("url")
		url = urlg.String()
		headerv, err := urlo.Get("headers")
		if err == nil {
			if headerv.IsObject() {
				headero := headerv.Object()
				for _, key := range headero.Keys() {
					value, _ := headero.Get(key)
					headers[key] = value.String()
				}
			}
		}
	}
	return url, headers
}

func httpGet(vm *vm.VM, call otto.FunctionCall) otto.Value {
	arg0 := call.Argument(0)
	url, headers := paramParse(arg0)
	resp, body, tls, err := comm.HttpGet(url, headers)
	var ottoReturn otto.Value
	if arg0.IsString() {
		var resultDisplay string
		if err != nil {
			resultDisplay = fmt.Sprintf("err %#v\n", err)
			ottoReturn, _ = otto.ToValue("") // scripts use length 0 as err indication
		} else {
			ottoReturn, _ = otto.ToValue(string(body))
			resultDisplay = fmt.Sprintf("%s %d bytes\n", resp.Status, len(body))
		}
		fmt.Printf("%s/%s http.get %s %s\n", vm.Owner, vm.Name, url, resultDisplay)
	} else {
		goResult := map[string]interface{}{}
		if err != nil {
			goResult["error"] = map[string]interface{}{
				"message": err.Error()}
		} else {
			fmt.Printf("%s/%s http.get %s %s %#v\n", vm.Owner, vm.Name, url, resp.Status)
			goResult["status"] = resp.StatusCode
			goResult["body"] = string(body)
			goResult["headers"] = resp.Header
			goTls := map[string]interface{}{}
			if tls != nil {
				tlsFill(goTls, tls)
			}
			goResult["tls"] = goTls
		}
		ottoReturn, err = vm.Js.ToValue(goResult)
		fmt.Printf("otto %#v %#v\n", ottoReturn, err)
	}
	return ottoReturn
}

func tlsFill(goTls map[string]interface{}, tls *tls.ConnectionState) {
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

func httpPost(vm *vm.VM, call otto.FunctionCall) otto.Value {
	arg0 := call.Argument(0)
	arg1 := call.Argument(1)
	url, headers := paramParse(arg0)
	body := ""
	if arg1.IsString() {
		body = call.Argument(1).String()
	} else if arg1.IsObject() {
		bodyo := arg1.Object()
		bodyBytes, err := json.Marshal(bodyo)
		if err != nil {
			fmt.Printf("httpPost json.Marshal err %#v\n", err)
		}
		body = string(bodyBytes)
		fmt.Printf("httpPost body %#v\n", bodyo.Class())
		fmt.Printf("httpPost body (%d) %#v\n", len(bodyBytes), bodyBytes)
		fmt.Printf("httpPost body %#v\n", body)
	} else {
		fmt.Printf("httpPost unknown body param %#v", arg1)
	}
	resp, err := comm.HttpPost(url, headers, strings.NewReader(body))
	var resultDisplay string
	var ottoStr otto.Value
	if err != nil {
		resultDisplay = fmt.Sprintf("err %#v\n", err)
		ottoStr, _ = otto.ToValue("")
	} else {
		resultDisplay = fmt.Sprintf("response body %#v bytes.\n", len(resp))
		ottoStr, _ = otto.ToValue(string(resp))
	}
	fmt.Printf("%s/%s http.post %s %s\n", vm.Owner, vm.Name, url, resultDisplay)
	return ottoStr
}
