build:
	go build -o bin/maahinen .cmd/maahinen/

run:
	go run ./cmd/maahinen

test:
	go test ./...

clean:
	rm -rf ./bin/maahinen

dev: build
	./bin/maahinen

tidy:
	go mod tidy