# BUILDER LAYER
FROM golang:1.15-alpine AS build-env

# add a user for security
RUN adduser -D -u 10000 builder
RUN mkdir /build/ && chown builder /build/
USER builder

# copy over the source code
WORKDIR /build/
COPY . /build/

# compile the server
RUN go build -o /build/sleep cmd/sleep-server/main.go

# SERVER LAYER
FROM alpine:3.12

# add a user for security
RUN adduser -D -u 10000 server
USER server

# copy the server binary
WORKDIR /
COPY --from=build-env /build/sleep /

# command
EXPOSE 8080
CMD ["/sleep"]
