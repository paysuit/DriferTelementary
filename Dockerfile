#
## ---- Build stage ----
#FROM golang:1.26-alpine as builder
#
#WORKDIR /build
##copy source to dest which is the working directory
#COPY . .
#
#RUN go mod download
#RUN go build -o ./bin
#
## ---- Run stage ----
#From alpine:3.20
#
#WORKDIR /app
#COPY --from=builder /build/bin /server
#
#EXPOSE 8080
#
#CMD ["./server"]



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

#LABEL authors="Madegwa"