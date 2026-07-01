FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/service ./services/blog/cmd

FROM gcr.io/distroless/static-debian12:nonroot
ENV GOMEMLIMIT=90MiB
ENV GOGC=50
COPY --from=builder /out/service /app/service
COPY configs/docker/blog.yaml /app/configs/blog.yaml
USER nonroot:nonroot
EXPOSE 5001
ENTRYPOINT ["/app/service"]
