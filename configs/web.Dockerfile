# Base image
FROM golang:1.18

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your application's source code
COPY . .

# Build your application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./budget-server ./cmd/budget-server/


# Expose the port your app runs on
EXPOSE 3000

# Run the application
CMD ["go", "run","./cmd/budget-server/"]
