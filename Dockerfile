FROM golang:1.26-bookworm AS builder

WORKDIR /src
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /src/backend
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/hrms-api ./cmd/api

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install --no-install-recommends --yes ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --uid 10001 --create-home app \
    && mkdir -p /app \
    && chown -R app:app /app
WORKDIR /app
COPY --from=builder /out/hrms-api /app/hrms-api
RUN chown app:app /app/hrms-api

USER app
ENV APP_ENV=production \
    HTTP_ADDR=:8080 \
    DATABASE_DRIVER=cloudbase_pg \
    TIMEZONE=Asia/Shanghai
EXPOSE 8080
CMD ["/app/hrms-api"]
