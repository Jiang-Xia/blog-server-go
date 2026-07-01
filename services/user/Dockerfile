FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/service ./services/monolith/cmd/user

FROM gcr.io/distroless/static-debian12:nonroot
ENV GOMEMLIMIT=70MiB
ENV GOGC=50
COPY --from=builder /out/service /app/service
COPY configs/docker/user.yaml /app/configs/user.yaml
USER nonroot:nonroot
EXPOSE 5002 50052
ENTRYPOINT ["/app/service"]
