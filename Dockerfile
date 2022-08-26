FROM golang:1.19-alpine as build

RUN apk add --update alpine-sdk ca-certificates upx && \
    echo "app:x:1000:1000::/_nonesuch:/bin/sodall" > /tmp/mini.passwd

WORKDIR /app
COPY . .

ARG logLevel
ENV LOGLVL=$logLevel

RUN make app

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /tmp/mini.passwd /etc/passwd
COPY --from=build /lib/ld-musl-x86_64.so.1 /lib/ld-musl-x86_64.so.1

USER 1000

CMD ["/app"]
COPY --from=build /app/app /app
