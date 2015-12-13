package comm

import (
  "fmt"
  "os"

  // message bus
  "github.com/gdamore/mangos"
  "github.com/gdamore/mangos/transport/tcp"
  "github.com/gdamore/mangos/protocol/bus"

)

func die(format string, v ...interface{}) {
  fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
  os.Exit(1)
}

func Start(pipe chan string) {
  fmt.Fprintln(os.Stdout, fmt.Sprintf("bus %s", "0.1"))

  var err error
  var url = "tcp://127.0.0.1:40899"

  fmt.Printf("bus on  %s\n", url)
  var bus_sock mangos.Socket
  bus_sock, err = bus.NewSocket()
  if err != nil {

  }
  bus_sock.AddTransport(tcp.NewTransport())
  err = bus_sock.Listen(url)
  if err != nil {
    die("can't listen on bus socket: %s", err.Error())
  }

  for i := 0; i < 10; i++ {
    bus_sock.Send([]byte("Hello"))
  }

  var msg []byte
  for {
    if msg, err = bus_sock.Recv(); err != nil {
      die("Cannot recv: %s", err.Error())
    }
    pipe <- string(msg)
  }

}
