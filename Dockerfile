# Use the official Golang base image
FROM golang:1.18 as builder

# Set the working directory
WORKDIR /app

# Copy the go.mod and go.sum files to the container
COPY go.mod ./

# Copy the entire project to the container
COPY . .

RUN go mod tidy

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o app ./cmd
#RUN make build

# Create a minimal runtime image
FROM alpine:latest

# Set the working directory to /app
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/app .

#RUN chmod +x /app/pipcas-lb


# Command to run the executable
CMD ["./app"]