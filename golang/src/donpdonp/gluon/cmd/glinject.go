package main

import (
  "os"
  "fmt"
  "strings"
  "encoding/json"

  "donpdonp/gluon/comm"
  "donpdonp/gluon/util"
)

func main() {
  my_uuid := util.Snowflake()
  bus := comm.PubsubFactory(my_uuid)

  bus.Start("localhost:6379")
  if len(os.Args) > 1 {
    msg := map[string]interface{}{"method":os.Args[1], "key":os.Getenv("GLUON_KEY")}
    msg["params"] = argsParse(os.Args)
    go bus.Loop()
    bus.Send(msg, func(pkt map[string]interface{}) {
      _json, err := json.Marshal(pkt["result"])
      if err != nil {
        errMsg := map[string]interface{}{"error":err.Error()}
        errJson, _ := json.Marshal(errMsg)
        fmt.Printf("%+v\n", errJson)
      } else {
        fmt.Printf("%+v\n", _json)
      }
    })
    <- bus.Pipe
  } else {
    fmt.Println("usage: ", os.Args[0], "<method name> --param_name=value")
  }

}

func argsParse(args []string) (map[string]string) {
  opts := map[string]string{}
  for _, element := range args {
    if len(element) > 2 && element[0:2] == "--" {
      eqidx := strings.Index(element, "=")
      key := element[2:eqidx]
      value := element[eqidx+1:len(element)]
      opts[key] = value
    }
  }
  return opts
}
