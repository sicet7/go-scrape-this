VERSION = "0.0.0"

main: dist
	@go build -ldflags "-X main.version=${VERSION}" main.go

install:
	@npm install

dist: install wasm
	@npm run build --application_version=${VERSION}

wasm: wasm_js
	@sh -c "GOOS=js GOARCH=wasm go build -ldflags \"-X main.version=${VERSION}\" -o public/module.wasm wasm.go"

wasm_js:
	@cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" "public/wasm_exec.js"

clean:
	@rm -rf node_modules main dist public/wasm_exec.js public/module.wasm