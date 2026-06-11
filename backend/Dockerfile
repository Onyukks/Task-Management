# ---- Build stage ----
FROM golang:1.26-alpine AS build
WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Static binary so the final image needs no libc.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

# ---- Run stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 appuser
USER appuser
WORKDIR /app
COPY --from=build /out/server /app/server

EXPOSE 8080
ENTRYPOINT ["/app/server"]
