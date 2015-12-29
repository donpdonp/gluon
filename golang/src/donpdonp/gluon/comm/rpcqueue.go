package comm

import (
  "github.com/streamrail/concurrent-map"
)

type Rpcqueue struct {
  q cmap.ConcurrentMap
}

func RpcqueueMake() Rpcqueue{
  return Rpcqueue{q: cmap.New() }
}