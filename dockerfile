# Stage 1: Build the Go binary
FROM golang:1.22.4 AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webhook .

# Stage 2: Create the final image
FROM instrumentisto/flutter:3.22.2

# Set the working directory in the container
WORKDIR /app

# Clone the repository
# RUN git clone https://github.com/africaudio/Mobile.git .

# # Install dependencies
# RUN flutter pub get

# Copy the pre-built webhook binary from the builder stage
COPY --from=builder /app/webhook /app/webhook

# Make the webhook binary executable
RUN chmod +x /app/webhook

# Expose the webhook port (assuming it uses port 8080)
EXPOSE 8080

# Set environment variables
# ENV MAILGUN_DOMAIN=your_mailgun_domain
# ENV MAILGUN_API_KEY=your_mailgun_api_key
# ENV EMAIL_SENDER=webhook@yourdomain.com
# ENV EMAIL_RECIPIENTS=recipient1@example.com,recipient2@example.com

# Start the webhook server
CMD ["/app/webhook"]