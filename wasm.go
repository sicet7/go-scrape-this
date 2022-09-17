package main

import "syscall/js"

var version string = "0.0.0"

func main() {
	done := make(chan struct{}, 0)
	js.Global().Set("GetWasmModuleVersion", js.FuncOf(GetVersion))
	<-done
}

func GetVersion(this js.Value, args []js.Value) interface{} {
	return version
}
