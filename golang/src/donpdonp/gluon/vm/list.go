package vm

type List struct {
  entries []VM
}

func (list *List) Add(new_vm VM) (bool, int) {
  pos := list.IndexOf(new_vm.Name)
  if pos == -1 {
    list.entries = append(list.entries, new_vm)
    return true, -1
  }
  return false, pos
}

func (list *List) Del(name string) (bool, int) {
  pos := list.IndexOf(name)
  if pos > -1 {
    list.entries = append(list.entries[:pos], list.entries[pos+1:]...)
    return true, pos
  }
  return false, -1
}

func (list *List) IndexOf(name string) (int) {
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
  ch := make(chan VM);
  go func () {
    for i := 0; i < len(list.entries); i++ {
        ch <- list.entries[i]
    }
    close(ch)
  } ();
  return ch
}

func (list *List) At(idx int) VM {
  return list.entries[idx]
}
