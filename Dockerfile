FROM golang:latest as builder

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download

COPY ./  ./

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
RUN go build \
    -o /go/bin/server \
    -ldflags '-s -w' \
    /go/src/cmd/terraform-server/*.go


FROM hashicorp/terraform:1.8.0 as runner

COPY --from=builder /go/bin/server /app/server
COPY config.yaml /app/

WORKDIR /app

ENTRYPOINT ["/app/server"]
