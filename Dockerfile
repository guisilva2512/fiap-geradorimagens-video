FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./cmd/${SERVICE}

FROM alpine:3.22

RUN apk add --no-cache ca-certificates ffmpeg tzdata

COPY --from=builder /out/service /usr/local/bin/service

ENTRYPOINT ["/usr/local/bin/service"]