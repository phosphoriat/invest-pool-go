FROM golang:1.26.0-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /server ./cmd
FROM alpine:3.23
RUN adduser -D -u 10001 app && mkdir /data && chown app:app /data
COPY --from=build /server /server
USER app
ENV HTTP_ADDR=:8080 DATABASE_PATH=/data/invest-pool.db
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s CMD wget -qO- http://127.0.0.1:8080/health || exit 1
ENTRYPOINT ["/server"]
