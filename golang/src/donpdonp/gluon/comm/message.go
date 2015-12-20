package comm

import (
  "fmt"
)

func Msg_check(msg map[string]interface{}, my_uuid string) (bool) {
  if msg["id"] != nil && msg["from"] != nil {
    if msg["method"] != nil || msg["result"] != nil || msg["error"] != nil {
      from := msg["from"].(string)
      if from == my_uuid {
        // drop my own msgs
        return false
      } else {
        return true
      }
    } else {
      fmt.Println("missing msg method/result/error!")
      return false
    }
  } else {
    fmt.Println("missing msg id/from!")
    return false
  }
}

