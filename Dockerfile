# Use the official Golang image with Alpine Linux as the base image for building
FROM golang:1.22-alpine as builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set the working directory inside the builder container
WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the builder container
COPY . .

# Print contents of /build directory for debugging
RUN ls -l /build

# Build the Go app with CGO enabled and musl tag for the correct architecture
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags musl -o email-service .

# Create a new stage for the final application image based on Alpine Linux
FROM alpine:3.18

# Set the working directory inside the container
WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /build/email-service ./email-service

# Copy environment configuration file
COPY --from=builder /build/.env ./

# Expose port if necessary (remove if not needed)
EXPOSE 8080

# Command to run the application
CMD ["./email-service"]
