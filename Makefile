jsonsum:
	mkdir -p bin
	go build -trimpath -o bin/jsonsum ./cmd/jsonsum
