FROM golang:alpine as builder

ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=1

RUN apk update && apk add --no-cache gcc linux-headers libc-dev

WORKDIR $GOPATH/src/malwatch/
COPY . .

WORKDIR $GOPATH/src/malwatch/cmd/malwatch
RUN go build -trimpath -ldflags="-w -s -extldflags=-static" -o /malwatch/

RUN /malwatch/malwatch install

FROM scratch

COPY --from=builder /malwatch/ /