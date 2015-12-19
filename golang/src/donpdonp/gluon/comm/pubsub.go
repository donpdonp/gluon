package comm

import (
  "fmt"
  "encoding/json"

  // redis
  "gopkg.in/redis.v3"
)

type Pubsub struct {
  client *redis.Client
  sock *redis.PubSub
  Pipe chan map[string]interface{}
  Connected chan bool
}

func PubsubFactory() (Pubsub) {
  new_bus := Pubsub{}
  new_bus.Pipe = make(chan map[string]interface{})
  new_bus.Connected = make(chan bool)
  return new_bus
}

func (comm *Pubsub) Start(addr string) {
  fmt.Printf("redis bus start %s\n", addr)
  comm.client = redis.NewClient(&redis.Options{Addr:addr})
  var err error
  comm.sock, err = comm.client.Subscribe("gluon")
  if err != nil {
    fmt.Println("subscribe err", err)
  }
}

func (comm *Pubsub) Loop() {
  for {
    msg, _ := comm.sock.ReceiveMessage()
    fmt.Println("<-"+msg.Payload)
    var pkt map[string]interface{}
    json.Unmarshal([]byte(msg.Payload), &pkt)
    comm.Pipe <- pkt
  }
}

func (comm *Pubsub) Send(msg map[string]interface{}) {
  line, _ := json.Marshal(msg)
  err := comm.client.Publish("gluon", string(line)).Err()
  if err != nil {
    fmt.Println("Send err", err)
  } else{
    fmt.Println("->"+string(line))
  }
}

