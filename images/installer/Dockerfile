FROM golang:1.24@sha256:2b1cbf278ce05a2a310a3d695ebb176420117a8cfcfcc4e5e68a1bef5f6354da AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o kwasm-node-installer ./cmd/node-installer
RUN /app/kwasm-node-installer -h

FROM scratch
COPY --from=builder /app/kwasm-node-installer /kwasm-node-installer

ENTRYPOINT ["/kwasm-node-installer"]
