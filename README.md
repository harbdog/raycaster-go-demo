# Building WASM for gh-pages

```bash
GOARCH=wasm GOOS=js go build -o raycaster-go-demo.wasm .
```

# Running local WASM test

```bash
go run github.com/hajimehoshi/wasmserve@latest .
```

- <http://localhost:8080>
