FROM golang:1.22-alpine as app-builder

ARG GITHUB_TOKEN
WORKDIR /go/src/app
COPY . .
RUN apk add git

ENV GOPRIVATE=github.com/kapetacom
RUN echo "machine github.com login oauth password ${GITHUB_TOKEN}" > ~/.netrc

# Static build required so that we can safely copy the binary over.
RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -o app

FROM alpine:latest as alpine-with-tz
RUN apk --no-cache add tzdata zip
WORKDIR /usr/share/zoneinfo
# -0 means no compression.  Needed because go's tz loader doesn't handle compressed data.
RUN zip -q -r -0 /zoneinfo.zip .

FROM scratch
# the test program:
COPY --from=app-builder /go/src/app/app /app
# the timezone data:
ENV ZONEINFO /zoneinfo.zip
COPY --from=alpine-with-tz /zoneinfo.zip /
# we need tls certificates in order to make https requests
# NB: this pulls directly from the upstream image, which already has the latest ca-certificates
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/app"]