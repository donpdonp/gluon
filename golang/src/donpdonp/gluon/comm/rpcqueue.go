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
