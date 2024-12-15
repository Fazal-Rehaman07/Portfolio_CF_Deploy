# Build stage
FROM golang:latest AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY go.mod ./

# Download third-party dependencies
RUN go mod download

COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o main . 

# Final stage
FROM public.ecr.aws/lambda/provided:al2023

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main ./main

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
ENTRYPOINT [ "./main" ]





