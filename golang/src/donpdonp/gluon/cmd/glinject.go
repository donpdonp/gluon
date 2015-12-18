package main

import (
  "os"
  "fmt"
  "strings"

  "donpdonp/gluon/comm"
)

func main() {
  bus, _ := comm.Factory()

  bus.Connect("tcp://127.0.0.1:40899")
  if len(os.Args) > 1 {
    msg := map[string]interface{}{"method":os.Args[1]}
    if len(os.Args) > 2 {
      msg["params"] = argsParse(os.Args)
    }
    bus.Send(msg)
    go bus.Loop()
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
