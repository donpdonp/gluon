package main

import "fmt"
import "strings"
import "time"
import	"crypto/tls"

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
		headers := map[string]string{}
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
		resp, body, tls, err := comm.HttpGet(urlc, headers)
		fmt.Printf("%s/%s http.get %s %s %#v\n", vm.Owner, vm.Name, urlc, resp.Status, 
			resp.Header.Get("Content-Type"))
		goResult := map[string]interface{}{}
		if err != nil {
			goResult["error"] = map[string]interface{}{
				"message": err.Error()}
		} else {
			goResult["status"] = resp.StatusCode
			goResult["body"] = string(body)
			goResult["headers"] = resp.Header
			goTls := map[string]interface{}{}
			if tls != nil {
				tlsFill(goTls, tls)
			}
			goResult["tls"] = goTls
		}
		ottoV, err = vm.Js.ToValue(goResult)
		fmt.Printf("otto %#v %#v\n", ottoV, err)
	}
	return ottoV
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
	urlc := call.Argument(0).String()
	body := call.Argument(1).String()
	headers := map[string]string{}
	body, err := comm.HttpPost(urlc, headers,
		strings.NewReader(body))
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
