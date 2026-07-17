# blog-service Docker 构建（学习路径；避免依赖 gcr.io/distroless）
FROM golang:1.26-alpine AS builder
WORKDIR /src
ENV GOTOOLCHAIN=auto
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/service ./services/blog/cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -H -u 65532 nonroot
ENV GOMEMLIMIT=90MiB
ENV GOGC=50
COPY --from=builder /out/service /app/service
COPY configs/docker/blog.yaml /app/configs/blog.yaml
USER nonroot
EXPOSE 5001 50051
ENTRYPOINT ["/app/service"]
