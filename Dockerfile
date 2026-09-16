FROM golang:1.26.4
WORKDIR /app
COPY go.mod go.sum /app
RUN go mod download
COPY . .
RUN go build -o wallet-app .
CMD ["./wallet-app"]