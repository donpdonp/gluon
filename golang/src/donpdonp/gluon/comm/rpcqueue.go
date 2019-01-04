package comm

import (
	"gopkg.in/streamrail/concurrent-map.v0"
)

type Rpcqueue struct {
	q cmap.ConcurrentMap
}

func RpcqueueMake() Rpcqueue {
	return Rpcqueue{q: cmap.New()}
}

func (rpcq *Rpcqueue) callbacksWaiting(name string)([]interface{}) {
  winners := []interface{}{}
  for obj := range rpcq.q.IterBuffered() {
    if name == obj.Key {
      winners = append(winners, obj.Val)
    }
  }
  return winners
}