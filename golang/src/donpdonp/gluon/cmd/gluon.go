package main;

import (
  "fmt"
  "donpdonp/gluon/comm"
)

func main() {
  bus := make(chan string)
  go comm.Start(bus)

  fmt.Println("bus started")

  for {
    msg := <-bus
    fmt.Println("main got: "+msg)
  }
}
