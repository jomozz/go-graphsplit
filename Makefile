.PHONY: build
build:
	go build -ldflags "-s -w" -o graphsplit ./cmd/graphsplit/


## FFI

ffi: 
	./extern/filecoin-ffi/install-filcrypto
.PHONY: ffi



