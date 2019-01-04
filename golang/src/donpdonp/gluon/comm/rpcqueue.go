package comm

import (
	"gopkg.in/streamrail/concurrent-map.v0"
)

type Rpcqueue struct {
	q cmap.ConcurrentMap
}

type Callback struct {
	Cb   func(map[string]interface{})
	Name string
}

func RpcqueueMake() Rpcqueue {
	return Rpcqueue{q: cmap.New()}
}

func (rpcq *Rpcqueue) callbacksWaiting(name string) []Callback {
	winners := []Callback{}
	for obj := range rpcq.q.IterBuffered() {
		if name == obj.Key {
			winners = append(winners, obj.Val.(Callback))
		}
	}
	return winners
}
