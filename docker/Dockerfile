FROM golang:1.25-bookworm

RUN apt-get update && apt-get install -y curl

WORKDIR /app

COPY . .

RUN make build

CMD ["make", "run"]