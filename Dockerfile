FROM golang@sha256:6600d9933c681cb38c13c2218b474050e6a9a288ac62bdb23aee13bc6dedce18 as builder

ARG version

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "10001" \    
    "appuser"

WORKDIR $GOPATH/src/github.com/fiskeben/meetjescraper/

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.Version=${version}" -o /go/bin/meetjescraper

FROM gcr.io/distroless/static

COPY --from=builder /go/bin/meetjescraper /go/bin/meetjescraper

ENTRYPOINT ["/go/bin/meetjescraper"]
