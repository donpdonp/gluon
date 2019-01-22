package comm

import "strings"
import "gopkg.in/streamrail/concurrent-map.v0"

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

func (rpcq *Rpcqueue) CallbacksWaiting(name string) []Callback {
	winners := []Callback{}
	for obj := range rpcq.q.IterBuffered() {
		val := obj.Val.(Callback)
		if name == val.Name {
			winners = append(winners, val)
		}
	}
	return winners
}

func (rpcq *Rpcqueue) Clear(name string) {
	winners := rpcq.CallbacksWaiting(name)
	for _, cb := range winners {
		rpcq.Finished(cb.Name)
	}
}

func (rpcq *Rpcqueue) CallbackNames() []string {
	winners := []string{}
	for obj := range rpcq.q.IterBuffered() {
		callback := obj.Val.(Callback)
		winners = append(winners, callback.Name)
	}
	return unique(winners)
}

func (rpcq *Rpcqueue) ToString() string {
	winners := []string{}
	for obj := range rpcq.q.IterBuffered() {
		callback := obj.Val.(Callback)
		winners = append(winners, callback.Name+"-"+obj.Key)
	}
	return strings.Join(winners, " ,")
}

func (rpcq *Rpcqueue) Count() int {
	return rpcq.q.Count()
}

func (rpcq *Rpcqueue) Finished(id string) {
	rpcq.q.Remove(id)
}

func unique(sslice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range sslice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
