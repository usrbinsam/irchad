FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /chadd .

FROM alpine:3

WORKDIR /

COPY --from=builder /chadd /chadd

USER nobody
EXPOSE 8888

ENTRYPOINT ["/chadd"]
