FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /app .
FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app -u 10001 app && mkdir -p /data /app/static && chown -R app:app /data
WORKDIR /app/static
COPY --from=build /app /app/server
COPY index.html donations.html supplier.html signup.html team.html admin.html camp.html scout.html cook.html runner.html builder.html command.html reports.html pointform.js i18n.js nav.js qr.js reports.js leaflet.css leaflet.js unidos-por-azuero.png ./
COPY report/ ./report/
USER app
ENV DB_PATH=/data/relief.db STATIC_DIR=/app/static LISTEN_ADDR=:8080
EXPOSE 8080
CMD ["/app/server"]
