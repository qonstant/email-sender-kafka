# Use the official Golang image with Alpine as the base image for building
FROM golang:1.22-alpine as builder

WORKDIR /app

# Install required packages
RUN apk update && apk --no-cache add build-base librdkafka-dev pkgconf musl-dev

# Go modules
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

# Environment
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Build
RUN go build -tags musl --ldflags '-linkmode external -extldflags "-static"' -o ./build/app ./...

FROM gcr.io/distroless/static-debian12 as runner

WORKDIR /app

# Copy the built application and .env file
COPY --from=builder /app/build/app /app/app
COPY --from=builder /app/.env /app/.env

ENTRYPOINT ["/app/app"]
