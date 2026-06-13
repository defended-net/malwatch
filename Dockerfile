FROM alpine:latest@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11 as builder-mimalloc

RUN apk add --no-cache cmake git build-base

ARG MIMALLOC_TAG=v3.1.5
ARG MIMALLOC_COMMIT=dfa50c37d951128b1e77167dd9291081aa88eea4

RUN git clone https://github.com/microsoft/mimalloc.git /mimalloc \
    && cd /mimalloc \
    && git checkout "${MIMALLOC_COMMIT}" \
    && commit=$(git rev-parse HEAD) \
    && [ "${commit}" = "${MIMALLOC_COMMIT}" ] || { echo "mimalloc commit mismatch: ${commit} != ${MIMALLOC_COMMIT}"; exit 1; } \
    && tagged=$(git rev-list -n 1 "refs/tags/${MIMALLOC_TAG}") \
    && [ "${tagged}" = "${MIMALLOC_COMMIT}" ] || { echo "mimalloc tag ${MIMALLOC_TAG} points to ${tagged}, expected ${MIMALLOC_COMMIT}"; exit 1; }

WORKDIR /mimalloc

RUN cmake . -Bbuild \
    -DMI_OVERRIDE=ON \
    -DCMAKE_C_FLAGS="-DMI_OPTION_LARGE_OS_PAGES_DEFAULT=0"

RUN cmake --build build

RUN cmake --install build

FROM golang:alpine@sha256:91eda9776261207ea25fd06b5b7fed8d397dd2c0a283e77f2ab6e91bfa71079d as builder-malwatch

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