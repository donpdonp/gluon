package comm

import (
  "fmt"
)

func Msg_check(msg map[string]interface{}, my_uuid string) (bool) {
  var ok bool
  if _, ok = msg["id"]; ok {
    var from interface{}
    if from, ok = msg["from"]; ok {
      if from == my_uuid {
        // drop my own msgs
        return false
      }
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

