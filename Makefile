.PHONY: build test test262 test262-quick clean

build:
	go build -o jsgo ./cmd/jsgo/
	go build -o test262runner ./cmd/test262runner/

test:
	go test ./...

test262: build
	./test262runner -dir test262 -v

test262-quick: build
	./test262runner -dir test262 -limit 100 -v

clean:
	rm -f jsgo test262runner
