
# ---- Build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY . .

RUN go mod download
RUN go build -o server ./`cmd`/main/main.go

# ---- Run stage ----
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /build/server .

EXPOSE 8080

CMD ["./server"]