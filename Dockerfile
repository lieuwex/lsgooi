FROM golang:1.13.5-alpine

# Install deps
RUN apk update && apk add git tzdata

WORKDIR /go/src/lsgooi/
COPY . .
RUN go build -o /bin/lsgooi

CMD ["/bin/lsgooi"]
