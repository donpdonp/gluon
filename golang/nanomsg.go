package main

import (
  "fmt"
  "os"

  "github.com/op/go-nanomsg"
)

func main() {
  var pub *nanomsg.PubSocket
  var err error
  if pub, err = nanomsg.NewPubSocket(); err != nil {
  }

  var url = "tcp://127.0.0.1:40899"
  pub.Bind(url)
  pub.Send([]byte("0123456789012345678901234567890123456789"), 0)

  var sub *nanomsg.SubSocket
  if sub, err = nanomsg.NewSubSocket(); err != nil {
    fmt.Fprintln(os.Stdout, fmt.Sprintf("sub sub err"))
  }
  sub.Subscribe("")
  fmt.Fprintln(os.Stdout, fmt.Sprintf("sub connecting"))
  _, err = sub.Connect(url)
  if err != nil {
    fmt.Fprintln(os.Stdout, fmt.Sprintf("sub connect err"))
  }
  fmt.Fprintln(os.Stdout, fmt.Sprintf("pub sending 2"))
  pub.Send([]byte("0123456789012345678901234567890123456789"), 0)
  fmt.Fprintln(os.Stdout, fmt.Sprintf("sub recv"))
  buf, err := sub.Recv(0)
  fmt.Fprintln(os.Stdout, fmt.Sprintf("sub recv blk"))
  if err != nil {
    fmt.Fprintln(os.Stdout, fmt.Sprintf("sub recv err"))
  }
  fmt.Fprintln(os.Stdout, fmt.Sprintf("sub say %s", buf))

}
