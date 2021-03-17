FROM golang:1.15-alpine as builder
RUN apk add --no-cache ca-certificates git

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o appbinary

FROM alpine as release
COPY --from=builder /go/src/appappbinary /appbinary
VOLUME /workspace
ENTRYPOINT ["/appbinary"]
