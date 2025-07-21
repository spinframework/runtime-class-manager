FROM alpine:3.22.1
WORKDIR /
COPY ./bin/manager /manager

ENTRYPOINT ["/manager"]
