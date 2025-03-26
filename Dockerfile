FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/app main/main.go

FROM alpine:3.17
WORKDIR /root/
COPY --from=builder /bin/app /bin/app

EXPOSE 8081
EXPOSE 50051

ENTRYPOINT ["/bin/app"]
