# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:alpine3.10  as builder

RUN apk --no-cache add build-base gcc
# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=1 go build -o main . 

FROM alpine:3.10

WORKDIR /app
RUN apk --no-cache add tzdata ca-certificates libc6-compat libgcc libstdc++

COPY --from=builder /app /app

# Expose port 8080 to the outside world
EXPOSE 8080

# ENV AWS_ACCESS_KEY_ID dummy
# ENV AWS_SECRET_ACCESS_KEY dummy
# ENV AWS_REGION us-east-1

# Command to run the executable
CMD ["./main"]