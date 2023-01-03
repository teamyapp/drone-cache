FROM golang:1.19-alpine AS builder

WORKDIR /workspace

COPY go.mod go.sum ./

RUN go mod download

RUN go mod verify

COPY . .

RUN go build -o bin/main main.go

FROM alpine:3.13 as production

WORKDIR /bin

COPY --from=builder /workspace/bin/main main

ENTRYPOINT ["/bin/main"]
