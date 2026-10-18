FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /ainews ./cmd/ainews

# ---- frontend build ----
FROM node:22-alpine AS frontend
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# ---- final image ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder  /ainews          ./ainews
COPY --from=frontend /web/dist        ./web/dist
COPY migrations/                      ./migrations/

EXPOSE 8080
ENV MODE=both
ENTRYPOINT ["/app/ainews"]
