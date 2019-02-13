FROM golang:1.11.4-alpine AS builder

# Install dep
RUN apk update && apk add git
RUN go get -u github.com/golang/dep/cmd/dep

# Install dependencies
COPY Gopkg.lock Gopkg.toml /go/src/stream-downloader/
WORKDIR /go/src/stream-downloader/
RUN dep ensure -vendor-only

COPY . .
RUN go build -o /bin/gooid

#####

FROM scratch AS runner

# Copy gooid
COPY --from=builder /bin/gooid /bin/gooid

CMD ["/bin/gooid"]
