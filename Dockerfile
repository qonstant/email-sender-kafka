# Use the official Golang image as the base image for building the application
FROM golang:1.22 as builder

# Set the working directory inside the container
WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the builder container
COPY . .

# Build the Go app
RUN go build -o email-service .

# Create a new stage for the final application image (based on Alpine Linux)
FROM alpine:3.18

# Install necessary packages
RUN apk add --no-cache libmagic ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /build/email-service ./email-service

# Copy environment file
COPY .env .env

# Expose port if necessary (if your app listens on a specific port, otherwise you can remove this line)
EXPOSE 8080

# Command to run your application
CMD ["./email-service"]
