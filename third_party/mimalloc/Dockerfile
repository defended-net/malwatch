FROM alpine:3.21@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d

RUN apk add --no-cache cmake git build-base

RUN git clone --depth 1 --branch v3.1.5 https://github.com/microsoft/mimalloc.git \
    && cd mimalloc \
    && test "$(git rev-parse HEAD)" = "dfa50c37d951128b1e77167dd9291081aa88eea4"

WORKDIR /mimalloc

RUN cmake . -Bbuild \
    -DMI_OVERRIDE=ON \
    -DCMAKE_C_FLAGS="-DMI_OPTION_LARGE_OS_PAGES_DEFAULT=0"

RUN cmake --build build

RUN cmake --install build