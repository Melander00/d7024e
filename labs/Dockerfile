# --- Build Stage ---
FROM golang:1.27.0 AS builder

WORKDIR /app

COPY go.mod ./
# COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/kademlia-bin /app/cmd/main.go

# --- Run Stage ---
FROM alpine:3.24 AS app

WORKDIR /app

RUN addgroup -g 1001 -S go && \
    adduser -S user -u 1001

COPY --from=builder --chown=user:go /app/kademlia-bin /app/kademlia-bin

RUN chmod +x /app/kademlia-bin

CMD ["/app/kademlia-bin"]
