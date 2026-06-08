FROM golang:1.26@sha256:68cb6d68bed024785b69195b89af7ac7a444f27791435f98647edff595aa0479 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o rcm-node-installer ./cmd/node-installer
RUN /app/rcm-node-installer -h

# Using busybox instead of scratch so that the nsenter utility is present, as used in restarter logic
FROM busybox:1.38@sha256:fd8d9aa63ba2f0982b5304e1ee8d3b90a210bc1ffb5314d980eb6962f1a9715d
COPY --from=builder /app/rcm-node-installer /rcm-node-installer

ENTRYPOINT ["/rcm-node-installer"]
