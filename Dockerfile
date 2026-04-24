FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/bangumi-pikpak .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/bangumi-pikpak /usr/local/bin/bangumi-pikpak
COPY example.config.json /app/example.config.json
VOLUME ["/app/data"]
CMD ["bangumi-pikpak", "-config", "/app/data/config.json", "-log", "/app/data/rss-pikpak.log"]

