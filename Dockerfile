FROM golang:1.15-alpine as builder
RUN apk add make binutils

COPY / /work
WORKDIR /work
RUN make

FROM alpine:3.12
COPY --from=builder /work/bin/kcm-watchdog /kcm-watchdog
USER root
ENTRYPOINT ["/kcm-watchdog","check"]
