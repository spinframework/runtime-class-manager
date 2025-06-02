FROM alpine:3.22.0

RUN apk add --no-cache curl bash tar
COPY download_shim.sh /download_shim.sh
CMD ["bash", "/download_shim.sh" ]
