package comm

import (
	"encoding/json"
	"fmt"

	// redis
	"gopkg.in/redis.v3"
)

type Pubsub struct {
	uuid      string
	Rpcq      Rpcqueue
	sclient   *redis.Client
	client    *redis.Client
	sock      *redis.PubSub
	Pipe      chan map[string]interface{}
	Connected chan bool
}

func PubsubFactory(uuid string, rpcq Rpcqueue) Pubsub {
	new_bus := Pubsub{uuid: uuid, Rpcq: rpcq}
	new_bus.Pipe = make(chan map[string]interface{})
	new_bus.Connected = make(chan bool)
	return new_bus
}

func (comm *Pubsub) Start(addr string) {
	comm.sclient = redis.NewClient(&redis.Options{Addr: addr})
	comm.client = redis.NewClient(&redis.Options{Addr: addr})
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

			if pkt["from"] != nil && pkt["from"].(string) == comm.uuid {
				// drop my own msgs
			} else {
				if pkt["id"] != nil {
					id := pkt["id"].(string)
					callback_obj, ok := comm.Rpcq.q.Get(id)
					if ok {
						callback := callback_obj.(Callback)
						callback.Cb(pkt)
					}
				}
				comm.Pipe <- pkt
			}
		}
	}
}

func (comm *Pubsub) Send(msg map[string]interface{}, callback *Callback) string {
	id := IdGenerate()
	msg["id"] = id
	msg["from"] = comm.uuid
	if callback != nil {
		comm.Rpcq.q.Set(id, *callback)
	}
	bytes, _ := json.Marshal(msg)
	line := string(bytes)
	err := comm.client.Publish("gluon", line).Err()
	if err != nil {
		fmt.Println("Send err", err)
	}
	return line
}
