FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
COPY cmd/eternalpose/main.go ./
COPY manga.json ./

RUN CGO_ENABLED=0 GOOS=linux go build main.go

CMD ["./main"]