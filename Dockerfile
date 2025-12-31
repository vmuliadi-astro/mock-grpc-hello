# build runs on build platform
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
WORKDIR /app

RUN apk add --no-cache protoc git ca-certificates file

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

COPY go.mod go.sum ./
RUN go mod download

COPY server ./server

RUN protoc -I server \
  --go_out=server --go_opt=paths=source_relative \
  --go-grpc_out=server --go-grpc_opt=paths=source_relative \
  server/helloworld.proto

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
  go build -trimpath -ldflags="-s -w" -o /out/greeter ./server && \
  chmod 755 /out/greeter && \
  file /out/greeter

# runtime uses target platform
FROM --platform=$TARGETPLATFORM gcr.io/distroless/static:nonroot
COPY --from=build /out/greeter /greeter
EXPOSE 50051
USER nonroot
ENTRYPOINT ["/greeter"]
