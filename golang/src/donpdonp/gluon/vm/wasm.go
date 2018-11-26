package vm
import "github.com/go-interpreter/wagon/exec"
import "github.com/go-interpreter/wagon/validate"
import "github.com/go-interpreter/wagon/wasm"

import "log"
import "strings"

func wasmfactory()(*exec.VM, error) {
  f := "wasm code"
  m, err := wasm.ReadModule(strings.NewReader(f), importer)
  if err != nil {
    log.Fatalf("could not read module: %v", err)
  }

  verify := true
  if verify {
    err = validate.VerifyModule(m)
    if err != nil {
      log.Fatalf("could not verify module: %v", err)
    }
  }

  if m.Export == nil {
    log.Fatalf("module has no export section")
  }

  return exec.NewVM(m)
}

func importer(name string) (*wasm.Module, error) {
  log.Printf("importer: %s\n", name);
  return nil, nil
}