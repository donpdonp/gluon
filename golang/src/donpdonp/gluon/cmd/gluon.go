package main;

import (
  "fmt"
  "donpdonp/gluon/comm"
)

func main() {
  go comm.Start()

  fmt.Println("bus started")
}
