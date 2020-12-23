FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM golang:latest as builder

RUN mkdir /IGVMultiBrowser
WORKDIR /IGVMultiBrowser
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o IGVMultiBrowser .

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /IGVMultiBrowser/IGVMultiBrowser .

COPY static /static
COPY templates /templates

ENTRYPOINT [ "/IGVMultiBrowser", "-c", "/config/config.yaml" ]