package comm

import (
  "fmt"
  "os"
  "encoding/json"

  // message bus
  "github.com/gdamore/mangos"
  "github.com/gdamore/mangos/transport/tcp"
  "github.com/gdamore/mangos/protocol/bus"

)

type Bus struct {
  sock mangos.Socket
  Pipe chan map[string]interface{}
}

func die(format string, v ...interface{}) {
  fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
  os.Exit(1)
}

func Factory() (Bus, error) {
  new_bus := Bus{}
  bus_sock, err := bus.NewSocket()
  new_bus.sock = bus_sock
  new_bus.Pipe = make(chan map[string]interface{})
  return new_bus, err
}

func (comm *Bus) Start(url string) {
  fmt.Fprintln(os.Stdout, fmt.Sprintf("bus %s", "0.1"))

  var err error

  fmt.Printf("bus on  %s\n", url)
  comm.sock.AddTransport(tcp.NewTransport())
  err = comm.sock.Listen(url)
  if err != nil {
    die("can't listen on bus socket: %s", err.Error())
  }

  var msg []byte
  for {
    if msg, err = comm.sock.Recv(); err != nil {
      die("Cannot recv: %s", err.Error())
    }
    jmsg := string(msg)
    fmt.Println("<-"+jmsg)
    var pkt map[string]interface{}
    json.Unmarshal(msg, &pkt)
    comm.Pipe <- pkt
  }

}

func (comm *Bus) Send(msg map[string]string) {
  line, _ := json.Marshal(msg)
  fmt.Println("->"+string(line))
  comm.sock.Send(line)
}