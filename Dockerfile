FROM golang:1.15-alpine as builder
RUN apk add --no-cache ca-certificates git

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o appbinary

FROM alpine as release
COPY --from=builder /go/src/app/appbinary /appbinary
COPY --from=builder /go/src/app/config.yaml /config.yaml
VOLUME /workspace
ENTRYPOINT ["/appbinary"]
