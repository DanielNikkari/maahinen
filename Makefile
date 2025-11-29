build:
	go build -o bin/maahinen ./cmd/maahinen

test:
	go test ./...

clean:
	rm -rf ./bin/maahinen

dev: build
	./bin/maahinen

run:
	@./bin/maahinen

tidy:
	go mod tidy