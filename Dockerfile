FROM alpine:latest as builder-mimalloc

RUN apk add --no-cache cmake git build-base

RUN git clone --depth 1 --branch v3.1.5 https://github.com/microsoft/mimalloc.git

WORKDIR /mimalloc

RUN cmake . -Bbuild \
    -DMI_OVERRIDE=ON \
    -DCMAKE_C_FLAGS="-DMI_OPTION_LARGE_OS_PAGES_DEFAULT=0"

RUN cmake --build build

RUN cmake --install build

FROM golang:alpine as builder-malwatch

ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=1

WORKDIR /mimalloc
COPY --from=builder-mimalloc /mimalloc /mimalloc

RUN apk update && apk add --no-cache gcc musl-dev linux-headers

WORKDIR $GOPATH/src/malwatch/
COPY . .

WORKDIR $GOPATH/src/malwatch/cmd/malwatch

RUN go build -trimpath --ldflags '-w -s -linkmode external -extldflags "-static -I/mimalloc/include -L/mimalloc/build -lmimalloc"' -o /malwatch/
RUN /malwatch/malwatch install

FROM scratch

COPY --from=builder-malwatch /malwatch/ /