VERSION = "0.0.0"

app: server/server
	@sh -c "cp server/server app"

server/server: server/dist
	@sh -c "cd server && go build -ldflags \"-X main.version=${VERSION}\" -o server main.go && cd .."

server/dist: client/dist
	@sh -c "cp -r client/dist server/dist"

client/node_modules:
	@npm install --prefix client

client/dist: client/node_modules client/public/wasm_exec.js client/public/module.wasm
	@npm run build --prefix client --application_version=${VERSION}

client/public/wasm_exec.js: wasm/wasm_exec.js
	@sh -c "cp wasm/wasm_exec.js client/public/wasm_exec.js"

client/public/module.wasm: wasm/module.wasm
	@sh -c "cp wasm/module.wasm client/public/module.wasm"

wasm/module.wasm:
	@sh -c "cd wasm && GOOS=js GOARCH=wasm go build -ldflags \"-X main.version=${VERSION}\" -o module.wasm main.go && cd .."

wasm/wasm_exec.js:
	@cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" "wasm/wasm_exec.js"

clean:
	@rm -rf server/dist client/node_modules client/dist client/public/wasm_exec.js client/public/module.wasm wasm/module.wasm wasm/wasm_exec.js server/server app