# Use an official Go runtime as a parent image
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY ./go.mod ./go.sum ./
RUN go mod download

# Copy the Go application source code into the container
COPY ./ ./

# Build the Go application
RUN CGO_ENABLED=0 go build -tags=jsoniter -o check-republic .

# Start a new stage for the minimal runtime container
FROM gcr.io/distroless/static-debian12

# Set the working directory inside the minimal runtime container
WORKDIR /app

# Copy the built binary from the builder container into the minimal runtime container
COPY --from=builder /app/check-republic . 
COPY models/regions.json ./models/regions.json

# Run your Go application and then sleep indefinitely
ENTRYPOINT ["/app/check-republic"]