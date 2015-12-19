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

func Factory() (Bus, error) {
  new_bus := Bus{}
  bus_sock, err := bus.NewSocket()
  bus_sock.AddTransport(tcp.NewTransport())
  new_bus.sock = bus_sock
  new_bus.Pipe = make(chan map[string]interface{})
  return new_bus, err
}

func (comm *Bus) Start(url string) {
  fmt.Fprintln(os.Stdout, fmt.Sprintf("bus %s", "0.1"))

  var err error

  fmt.Printf("bus on  %s\n", url)
  err = comm.sock.Listen(url)
  if err != nil {
    fmt.Println("can't listen on bus socket: %s", err.Error())
  }
}

func (comm *Bus) Loop() {
  var msg []byte
  var err error
  for {
    if msg, err = comm.sock.Recv(); err != nil {
      fmt.Println("Cannot recv: %s", err.Error())
    }
    jmsg := string(msg)
    fmt.Println("<-"+jmsg)
    var pkt map[string]interface{}
    json.Unmarshal(msg, &pkt)
    comm.Pipe <- pkt
  }
}

func (comm *Bus) Connect(url string) {
  fmt.Println("bus connect", url)
  err := comm.sock.Dial(url)
  if err != nil {
    fmt.Println("can't listen on bus socket: %s", err.Error())
  }
}

func (comm *Bus) Send(msg map[string]interface{}) {
  line, _ := json.Marshal(msg)
  err := comm.sock.Send(line)
  if err != nil {
    fmt.Println("Send err", err)
  } else{
    fmt.Println("->"+string(line))
  }
}
