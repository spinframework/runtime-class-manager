FROM golang:1.26@sha256:595c7847cff97c9a9e76f015083c481d26078f961c9c8dca3923132f51fe12f1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o rcm-node-installer ./cmd/node-installer
RUN /app/rcm-node-installer -h

# Using busybox instead of scratch so that the nsenter utility is present, as used in restarter logic
FROM busybox:1.37@sha256:b3255e7dfbcd10cb367af0d409747d511aeb66dfac98cf30e97e87e4207dd76f
COPY --from=builder /app/rcm-node-installer /rcm-node-installer

ENTRYPOINT ["/rcm-node-installer"]
