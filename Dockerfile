# -=-=-=-=-=-=- Compile Image -=-=-=-=-=-=-

FROM golang:1 AS stage-compile

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./... && CGO_ENABLED=0 GOOS=linux go build ./cmd/go-start

# -=-=-=-=- Final Distroless Image -=-=-=-=-

# hadolint ignore=DL3007
FROM gcr.io/distroless/static-debian12:latest as stage-final

COPY --from=stage-compile /go/src/app/go-start /
CMD ["/go-start"]