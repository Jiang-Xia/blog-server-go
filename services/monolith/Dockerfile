# 单体 Docker 构建：多阶段编译 → alpine 运行（本地试验；避免依赖 gcr.io/distroless）
# go.mod 要求 >=1.26.4；国内构建默认走 goproxy.cn
FROM golang:1.26-alpine AS builder
WORKDIR /src
ENV GOTOOLCHAIN=auto
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/service ./services/monolith/cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -H -u 65532 nonroot
ENV GOMEMLIMIT=128MiB
ENV GOGC=50
ENV CONFIG_PATH=/app/configs/monolith.yaml
COPY --from=builder /out/service /app/service
COPY configs/docker/monolith.yaml /app/configs/monolith.yaml
USER nonroot
EXPOSE 8000
ENTRYPOINT ["/app/service"]
