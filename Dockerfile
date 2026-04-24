FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY third_party/ ./third_party/
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/bangumi-pikpak .

FROM node:22-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi
COPY frontend/ ./
RUN npm run build

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/bangumi-pikpak /usr/local/bin/bangumi-pikpak
COPY --from=frontend /src/frontend/dist /app/frontend/dist
COPY .env.example /app/.env.example
VOLUME ["/app/data"]
CMD ["bangumi-pikpak", "-config", "/app/data/.env", "-log", "/app/data/rss-pikpak.log"]

