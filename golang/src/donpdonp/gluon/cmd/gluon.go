package main;

import (
  "fmt"
  "donpdonp/gluon/comm"
)

func main() {
  bus, _ := comm.Factory()

  go bus.Start()

  fmt.Println("bus started")

  bus.Send(map[string]string{"a":"b"})
  for {
    msg := <-bus.Pipe
    fmt.Println("main got: "+msg)
  }
}
