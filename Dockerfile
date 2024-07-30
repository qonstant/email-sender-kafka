FROM golang:1.22-alpine as builder

WORKDIR /app

# Packages
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
RUN go build -tags musl --ldflags '-linkmode external -extldflags "-static"' -o ./build/...

FROM gcr.io/distroless/static-debian12 as runner

COPY --from=builder ["/app/build/", "/"]

EXPOSE 8080

ENTRYPOINT [...]