
# ---- Build stage ----
FROM golang:1.26-alpine as builder

WORKDIR /build
#copy source to dest which is the working directory
COPY . .

RUN go mod download
RUN go build -o ./bin

# ---- Run stage ----
From alpine:3.20

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
