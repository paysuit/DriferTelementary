
# ---- Build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY main.go .

RUN go build -o server main.go

# ---- Run stage ----
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]

LABEL authors="Madegwa"

ENTRYPOINT ["top", "-b"]