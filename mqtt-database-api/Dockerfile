# Start from the latest Golang base image for ARM
FROM --platform=${BUILDPLATFORM} golang:1.19 as builder

LABEL maintainer="Adam Hofman <adam.hofman1@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY ./mqtt-database-api ./mqtt-database-api
COPY ./database ./database
COPY ./product-data ./product-data
COPY ./mqtt-client ./mqtt-client

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -o main ./mqtt-database-api/cmd/main.go

# Start a new stage from scratch
FROM --platform=${BUILDPLATFORM} scratch

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# This container exposes port 1883 to the outside world
EXPOSE 1883

# Run the binary program produced by ‘go install’
CMD ["./main"]
