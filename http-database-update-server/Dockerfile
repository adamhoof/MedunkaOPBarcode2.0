# Start from the latest Golang base image for ARM
FROM --platform=linux/arm64 golang:latest as builder

# Add Maintainer Info
LABEL maintainer="Adam Hofman <adam.hofman1@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY ./http-database-update-server ./http-database-update-server
COPY ./database ./database
COPY ./product-data ./product-data

# Build the Go app for ARM architecture
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -o main ./http-database-update-server/cmd/main.go

# Start a new stage from scratch
FROM --platform=linux/arm64 alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Command to run the executable
CMD ["./main"]
