FROM golang:1.23@sha256:927112936d6b496ed95f55f362cc09da6e3e624ef868814c56d55bd7323e0959 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o rcm-node-installer ./cmd/node-installer
RUN /app/rcm-node-installer -h

FROM scratch
COPY --from=builder /app/rcm-node-installer /rcm-node-installer

ENTRYPOINT ["/rcm-node-installer"]
