FROM golang:1.22.3-bullseye as base
WORKDIR /app

RUN apt-get update && apt-get install -y curl clang gcc llvm make libbpf-dev

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading
# them in subsequent builds if they change
ENV GOPROXY https://goproxy.cn
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg \
    go mod download && go mod verify

FROM base as builder
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    make build-sdk

FROM debian:bullseye-slim AS runner
WORKDIR /app
ENV TZ=Asia/Shanghai
COPY --from=builder /app/sdk-auto.yml /app/
COPY --from=builder /app/originx-sdk-auto /app/
CMD ["/app/originx-sdk-auto", "--config=/app/sdk-auto.yml"]