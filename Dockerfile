FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o go_app ./cmd
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
CMD ["./go_app"]
