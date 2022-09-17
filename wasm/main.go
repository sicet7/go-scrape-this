package main

import (
	"go-scrape-this/wasm/utils"
	"syscall/js"
)

var version string = "0.0.0"
var shutdown chan bool

func init() {
	shutdown = make(chan bool)
}

func main() {
	js.Global().Set("GetWasmModuleVersion", utils.MakePromiseFunction(GetVersion))
	<-shutdown
	println("WASM module is shutting down.")
}

func GetVersion(this js.Value, args []js.Value, returner utils.Returner) {
	returner.Resolve.Invoke(map[string]interface{}{
		"version": version,
		"args 0":  args[0],
		"this":    this,
		"error":   nil,
	})
}
