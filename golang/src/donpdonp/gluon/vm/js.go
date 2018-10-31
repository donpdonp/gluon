package vm

import "fmt"

func (vm *VM) EvalJs(js_code string) (string, error) {
  result, err := vm.Js.Run(js_code)
  if err != nil {
    fmt.Println("evaljs Run failed", err, vm.Js.Context().Stacktrace, js_code)
    return "", err
  } else {
    fmt.Printf("evaljs JSON.stringify %#v\n", result)
    otto_json, err := vm.Js.Call("JSON.stringify", nil, result)
    json := ""
    if err != nil {
      fmt.Printf("evaljs JSON.stringify err: %#v\n", err)
    } else {
      thing, _ := otto_json.Export()
      if thing != nil {
        json = thing.(string)
        fmt.Printf("evaljs JSON.stringify good: %#v\n", json)
      }
    }
    return json, nil //descriptor_value json
  }
}

// ug, hack to support first time setup that returns a function
func (vm *VM) FirstEvalJs(js_code string) (string, error) {
  src, err := vm.Js.Compile("", js_code)
  if err != nil {
    fmt.Println("js compile failed!", err)
    return "", err
  } else {
    fmt.Println("js compile good!")
    setup, err := vm.Js.Run(src)
    if err != nil {
      fmt.Println("eval failed", err, vm.Js.Context().Stacktrace)
      return "", err
    } else {
      descriptor_value, err := setup.Call(setup)
      if err != nil {
        fmt.Println("js func setup eval fail %v", err)
        return "", err
      } else {
        otto_json, _ := vm.Js.Call("JSON.stringify", nil, descriptor_value)
        json, _ := otto_json.Export()
        return json.(string), nil //descriptor_value json
      }
    }
  }
}
