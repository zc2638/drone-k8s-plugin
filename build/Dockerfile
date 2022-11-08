FROM golang:1.19 as builder

ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct

WORKDIR /work
ADD . .
RUN make build

FROM alpine:3.10 as alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk add -U --no-cache ca-certificates tzdata

FROM alpine:3.10

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /work/plugin /usr/local/bin/plugin

WORKDIR /work
CMD ["plugin"]
