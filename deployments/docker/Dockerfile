####### STAGE 'builder'

FROM golang:1-alpine as builder

ARG BIN_NAME=server
ARG BUILD_DIR=/go/src/${BIN_NAME}

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN apk --no-cache add binutils make curl

ARG build_version="NOT_SET_DURING_BUILD_TIME"
ENV BUILD_VERSION=$build_version

WORKDIR ${BUILD_DIR}

COPY ./go.mod ./go.sum ./Makefile ./

RUN make mod-download

COPY ./cmd/ cmd/
COPY ./pkg/ pkg/
COPY ./internal/ ./internal/

RUN make build-sources BUILD_VERSION=$BUILD_VERSION OUTPUT_PATH=${BUILD_DIR} TARGET=${BIN_NAME}
RUN strip ${BIN_NAME}

# When root CA required.
RUN curl https://curl.se/ca/cacert.pem -o ca-bundle.pem

####### STAGE

FROM scratch

WORKDIR /etc/ssl/
COPY --from=builder /go/src/server/ca-bundle.pem .

WORKDIR /app
COPY --from=builder /go/src/server/server /app

ENTRYPOINT ["/app/server"]
