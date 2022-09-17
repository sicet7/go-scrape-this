package utils

import "syscall/js"

type Returner struct {
	Resolve js.Value
	Reject  js.Value
}

func MakePromiseFunction(f func(js.Value, []js.Value, Returner)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		handler := js.FuncOf(func(handlerThis js.Value, handlerArgs []js.Value) interface{} {
			p := Returner{
				Resolve: handlerArgs[0],
				Reject:  handlerArgs[1],
			}
			go f(this, args, p)
			return nil
		})
		promiseConstructor := js.Global().Get("Promise")
		return promiseConstructor.New(handler)
	})
}
