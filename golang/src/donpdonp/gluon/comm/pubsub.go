package comm

import (
  "fmt"
  "encoding/json"

  // redis
  "gopkg.in/redis.v3"
  "github.com/satori/go.uuid"

)

var (
  rpcq = RpcqueueMake()
)

type Pubsub struct {
  uuid string
  sclient *redis.Client
  client *redis.Client
  sock *redis.PubSub
  Pipe chan map[string]interface{}
  Connected chan bool
}


func PubsubFactory(uuid string) (Pubsub) {
  new_bus := Pubsub{uuid: uuid}
  new_bus.Pipe = make(chan map[string]interface{})
  new_bus.Connected = make(chan bool)
  return new_bus
}

func (comm *Pubsub) Start(addr string) {
  fmt.Printf("redis bus start %s\n", addr)
  comm.sclient = redis.NewClient(&redis.Options{Addr:addr})
  comm.client = redis.NewClient(&redis.Options{Addr:addr})
  var err error
  comm.sock, err = comm.sclient.Subscribe("gluon")
  if err != nil {
    fmt.Println("subscribe err", err)
  }
}

func (comm *Pubsub) Loop() {
  for {
    msg, err := comm.sock.ReceiveMessage()
    if err != nil {
      fmt.Println("<- receive err", err)
    } else {
      var pkt map[string]interface{}
      json.Unmarshal([]byte(msg.Payload), &pkt)

      if pkt["from"].(string) == comm.uuid {
        // drop my own msgs
      } else {
        if pkt["id"] != nil {
          callback, ok := rpcq.q.Get(pkt["id"].(string))
          if ok {
            rpcq.q.Remove(pkt["id"].(string))
            callback.(func(map[string]interface{}))(pkt)
          }
        }

        comm.Pipe <- pkt
      }
    }
  }
}

func (comm *Pubsub) Send(msg map[string]interface{}, callback func(map[string]interface{})) {
  msg["id"] = uuid.NewV4().String()[0:8]
  msg["from"] = comm.uuid
  if callback != nil {
    rpcq.q.Set(msg["id"].(string), callback)
  }
  line, _ := json.Marshal(msg)
  err := comm.client.Publish("gluon", string(line)).Err()
  if err != nil {
    fmt.Println("Send err", err)
  } else{
    fmt.Println("->"+string(line))
  }
}

