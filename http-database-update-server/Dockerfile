# Start from the latest Golang base image for ARM
FROM --platform=${BUILDPLATFORM} golang:1.19 as builder

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
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -a -installsuffix cgo -o main ./http-database-update-server/cmd/main.go

# Start a new stage from scratch
FROM --platform=${BUILDPLATFORM} alpine:3.10


WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Command to run the executable
CMD ["./main"]

#CMD ["/bin/sh"]

#sudo docker  build -t test-http-database-update-server -f ./http-database-update-server/Dockerfile .
#sudo docker run -it -e CSV_OUTPUT_PATH=/app/output.csv --env-file ./env.list --network=whales-network -d test-http-database-update-server
