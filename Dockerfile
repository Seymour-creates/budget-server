# Use the official Go image as a builder stage to compile the binary.
FROM golang:1.18 as builder

# Set the working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum to download the dependencies.
COPY go.mod ./
COPY go.sum ./

# Download Go modules.
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Build the Go app - adjust the path according to your main package location.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o budget-server ./cmd/budget-server

# Use a small image to create a minimal final image.
FROM alpine:latest

# Install CA certificates for HTTPS requests.
#RUN apk --no-cache add ca-certificates

# Create a non-root user and switch to it for security purposes.
#RUN adduser -D user
#USER user

# Set the working directory in the container.
#WORKDIR /app

# Copy the binary from the builder stage to the final image.
COPY --from=builder /app/budget-server .

# Copy other necessary files like dev.env, configs, etc., if needed.
# Be careful with sensitive data in dev.env - it might be better to use environment variables.
#COPY --from=builder /app/.env ./

# Expose the port the server listens on.
EXPOSE 3000

# Command to run the executable.
CMD ["./budget-server"]
