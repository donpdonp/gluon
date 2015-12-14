package main;

import (
  "fmt"
  "donpdonp/gluon/comm"
)

func main() {
  bus, _ := comm.Factory()

  go bus.Start("tcp://127.0.0.1:40899")

  fmt.Println("bus started")

  bus.Send(map[string]string{"a":"b"})

  for {
    msg := <-bus.Pipe
    method := msg["method"].(string)
    fmt.Println("method: "+method)

    switch method {
    case "vm.add":
      fmt.Println("adding")
    }
  }
}
