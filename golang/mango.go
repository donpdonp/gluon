package main

import (
  "fmt"
  "os"
  "time"

  // javascript interpreter
  "github.com/robertkrimen/otto"

  // message bus
  "github.com/gdamore/mangos"
  "github.com/gdamore/mangos/protocol/pub"
  "github.com/gdamore/mangos/protocol/sub"
  "github.com/gdamore/mangos/transport/tcp"

)

func die(format string, v ...interface{}) {
  fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
  os.Exit(1)
}

func main() {
  fmt.Fprintln(os.Stdout, fmt.Sprintf("gluon %s", "0.1"))

  vm := otto.New()
  vm.Run(`
      abc = 2 + 2;
      console.log("The value of abc is " + abc); // 4
  `)

  var err error
  var url = "tcp://127.0.0.1:40899"

  fmt.Fprintln(os.Stdout, fmt.Sprintf("pub"))
  var pub_sock mangos.Socket
  pub_sock, err = pub.NewSocket()
  if err != nil {
    die("can't get new pub socket: %s", err)
  }
  pub_sock.AddTransport(tcp.NewTransport())
  err = pub_sock.Listen(url)
  if err != nil {
    die("can't listen on pub socket: %s", err.Error())
  }
  err = pub_sock.Send([]byte(time.Now().Format(time.ANSIC)))

  fmt.Fprintln(os.Stdout, fmt.Sprintf("sub"))
  var sub_sock mangos.Socket
  sub_sock, err = sub.NewSocket()
  if err != nil {

  }
  sub_sock.AddTransport(tcp.NewTransport())
  err = sub_sock.Dial(url)
  if err != nil {
    die("can't dial on sub socket: %s", err.Error())
  }

  err = sub_sock.SetOption(mangos.OptionSubscribe, []byte(""))
  if err != nil {
    die("cannot subscribe: %s", err.Error())
  }

  var msg []byte
  for {
    if msg, err = sub_sock.Recv(); err != nil {
      die("Cannot recv: %s", err.Error())
    }
    fmt.Printf("CLIENT(%s): RECEIVED %s\n", "name", string(msg))
  }

}
