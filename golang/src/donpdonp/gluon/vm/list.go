package vm

import (
	"errors"
)

type List struct {
	entries  []VM
	Backchan chan map[string]interface{}
}

func ListFactory() List {
	thing := List{}
	thing.Backchan = make(chan map[string]interface{})
	return thing
}

func (list *List) Add(new_vm VM) (bool, int) {
	pos := list.IndexOf(new_vm.Name)
	if pos == -1 {
		list.entries = append(list.entries, new_vm)
		return true, -1
	}
	return false, pos
}

func (list *List) Del(name string) (string, error) {
	pos := list.IndexOf(name)
	if pos > -1 {
		winner := list.At(pos)
		list.entries = append(list.entries[:pos], list.entries[pos+1:]...)
		return winner.Url, nil
	}
	return "", errors.New("missing")
}

func (list *List) IndexOf(name string) int {
	for i, vm := range list.entries {
		if vm.Name == name {
			return i
		}
	}
	return -1
}

// crazy go iterator
// http://ewencp.org/blog/golang-iterators/
func (list *List) Range() <-chan VM {
	ch := make(chan VM)
	go func() {
		for i := 0; i < len(list.entries); i++ {
			ch <- list.entries[i]
		}
		close(ch)
	}()
	return ch
}

func (list *List) At(idx int) VM {
	return list.entries[idx]
}

func (list *List) Size() int {
	return len(list.entries)
}
