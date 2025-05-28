# Stage 1: Build the application
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main -ldflags="-s -w" .

# Stage 2: Create the final, smaller image
FROM alpine:latest

# --- TAMBAHKAN BARIS INI ---
# Install timezone data
RUN apk add --no-cache tzdata
# ---------------------------

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 3000

CMD ["./main"]