FROM alpine:3.22.0
WORKDIR /
COPY ./bin/manager /manager

ENTRYPOINT ["/manager"]
