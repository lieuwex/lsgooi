FROM golang:1.11.4-alpine

# Install deps
RUN apk update && apk add git tzdata
RUN go get -u github.com/golang/dep/cmd/dep

# Install dependencies
COPY Gopkg.lock Gopkg.toml /go/src/lsgooi/
WORKDIR /go/src/lsgooi/
RUN dep ensure -vendor-only

COPY . .
RUN go build -o /bin/lsgooi

CMD ["/bin/lsgooi"]
