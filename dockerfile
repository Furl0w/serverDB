FROM golang:1.10-alpine as builder

COPY Gopkg.lock Gopkg.toml /go/src/serverMongoDB/
WORKDIR /go/src/serverMongoDB
RUN apk add git
RUN apk add dep
RUN dep ensure -vendor-only
COPY app/ /go/src/serverMongoDB/app
COPY db/ /go/src/serverMongoDB/db
WORKDIR /go/src/serverMongoDB/app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o serverDB
FROM scratch

COPY --from=builder /go/src/serverMongoDB/app/serverDB /app/serverDB
CMD ["/app/serverDB"]