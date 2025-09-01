FROM alpine:latest

RUN apk add --no-cache cmake git build-base

RUN git clone --depth 1 --branch v3.1.5 https://github.com/microsoft/mimalloc.git

WORKDIR /mimalloc

RUN cmake . -Bbuild \
    -DMI_OVERRIDE=ON \
    -DCMAKE_C_FLAGS="-DMI_OPTION_LARGE_OS_PAGES_DEFAULT=0"

RUN cmake --build build

RUN cmake --install build