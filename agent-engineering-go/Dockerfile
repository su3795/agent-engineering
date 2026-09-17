# syntax=docker/dockerfile:1.6
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/agent-engineering ./cmd/agent-engineering

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /out/agent-engineering /app/
USER app
ENV GIN_MODE=release \
    OTEL_ENABLED=false
ENTRYPOINT ["/app/agent-engineering"]
CMD ["verify"]
