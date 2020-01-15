FROM golang:1.13-alpine3.11 AS builder
RUN apk --no-cache add build-base git bzr gcc
ADD . /src
RUN cd /src/check && CGO_ENABLED=0 GOOS=linux go build -o check
RUN cd /src/in && CGO_ENABLED=0 GOOS=linux go build -o in

# final stage
FROM alpine:3.11
COPY --from=builder /src/check/check /opt/resource/check
COPY --from=builder /src/in/in /opt/resource/in
