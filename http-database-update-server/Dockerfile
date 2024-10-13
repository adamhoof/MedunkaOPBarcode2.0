# Start from the latest Golang base image for ARM
FROM --platform=linux/arm64 golang:latest as builder

# Add Maintainer Info
LABEL maintainer="Adam Hofman <adam.hofman1@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Check Go version and environment (for debugging)
RUN go version && go env

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY ./http-database-update-server ./http-database-update-server
COPY ./database ./database
COPY ./product-data ./product-data

# Build the Go app for ARM architecture
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -o main ./http-database-update-server/cmd/main.go

# Start a new stage from scratch
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Command to run the executable
CMD ["./main"]

#CMD ["/bin/sh"]

#sudo docker  build -t test-http-database-update-server -f ./http-database-update-server/Dockerfile .
#sudo docker run -it -e CSV_OUTPUT_PATH=/app/output.csv --env-file ./env.list --network=whales-network -d test-http-database-update-server