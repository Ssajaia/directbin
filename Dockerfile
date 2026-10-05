FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY main.go ./

RUN mkdir -p /out && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/directbin .

FROM alpine:3.21

RUN addgroup -S directbin && \
    adduser -S -G directbin directbin && \
    mkdir /data && \
    chown directbin:directbin /data

COPY --from=build /out/directbin /directbin

USER directbin
WORKDIR /

ENV ADDR=:8080

EXPOSE 8080
VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/directbin"]