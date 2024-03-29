# Base image
FROM golang:1.21

# Set the working directory inside the container
WORKDIR /app

RUN apt-get clean && apt-get update && apt-get install -y \
    ca-certificates \
    default-mysql-client \
&& rm -rf /var/lib/apt/lists/*

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your application's source code
COPY . .

# Make the wait-for-mysql.sh script executable
RUN chmod +x configs/wait-for-mysql.sh

# Build your application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./budget-server ./cmd/budget-server/


# Expose the port your app runs on
EXPOSE 3000

# Run the application
CMD ["/bin/sh", "-c", "./configs/wait-for-mysql.sh db && go run ./cmd/budget-server/"]
