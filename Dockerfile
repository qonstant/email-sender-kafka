# Use the official Golang image with Alpine Linux as the base image for building
FROM golang:1.22-alpine as builder

RUN apk add --no-progress --no-cache gcc musl-dev
WORKDIR /build
COPY . .
RUN go mod download

RUN go build -tags musl -ldflags '-extldflags "-static"' -o /build/main

FROM scratch
WORKDIR /app
COPY --from=builder /build/main .
COPY .env .env
ENTRYPOINT ["/app/main"]
