FROM alpine:3.21.3
WORKDIR /
COPY ./bin/manager /manager

ENTRYPOINT ["/manager"]
