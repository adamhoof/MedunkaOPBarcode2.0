# Start from the latest Golang base image
FROM golang:latest as builder

# Add Maintainer Info
LABEL maintainer="Adam Hofman <adam.hofman1@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY ./cli-control-app .
COPY ./file-parser ./file-parser

# Install mdb-tools
RUN apt-get update && apt-get install -y mdbtools

# This will compile the application and produce /app/main
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Start a new stage from ubuntu:latest
FROM ubuntu:latest

# Install mdb-tools
RUN apt-get update && apt-get install -y mdbtools && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/file-parser ./file-parser

# Command to run the executable
#CMD ["./main"]

CMD ["/bin/bash"]

#sudo docker run -it --network=whales-network -v /home/adamhoof/Desktop/:/app/mdb -e MDB_PATH=/app/mdb/67668305_2022.mdb -e CSV_OUTPUT_PATH=/app/output.csv -e SHELL_MDB_FILE_PARSER_PATH=./file-parser/mdb_file_parser.sh --env-file ./env.list test-cli-control
#sudo docker  build -t test-cli-control -f ./cli-control-app/Dockerfile .