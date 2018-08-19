package comm

import (
	"fmt"
	"gopkg.in/satori/go.uuid.v1"
)

func IdGenerate() string {
	return uuid.NewV4().String()[0:8]
}

func Msg_check(msg map[string]interface{}) bool {
	var ok bool
	if _, ok = msg["id"]; ok {
		if _, ok = msg["from"]; ok {
			var allok = false
			if _, ok = msg["method"]; ok {
				allok = true
			}
			if _, ok = msg["result"]; ok {
				allok = true
			}
			if _, ok = msg["error"]; ok {
				allok = true
			}
			if !allok {
				fmt.Println("missing msg method/result/error!")
				return false
			}
			return allok
		} else {
			fmt.Println("missing msg from!")
			return false
		}
	} else {
		fmt.Println("missing msg id!")
		return false
	}
}
