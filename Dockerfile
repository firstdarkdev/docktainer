# Stage 1: Build the GO web and build server
FROM golang:1.23-alpine AS builder

# Install dependencies
RUN apk add --no-cache git curl bash

# Set up the working directory
WORKDIR /app
COPY ./server/. .

# Build the Go application
RUN go build -o docktainer

# Stage 2: Final container with Retype, Git and the Webserver
FROM mcr.microsoft.com/dotnet/sdk:9.0-alpine-amd64

# Install required tools
RUN apk add --no-cache \
    git \
    curl \
    bash \
    su-exec

# Install and Configure Retype
RUN dotnet tool install retypeapp --tool-path /bin
ENV RETYPE_DEFAULT_HOST="0.0.0.0"

# Set up Working Directory, and copy Web/Build Server from Builder
WORKDIR /app
COPY --from=builder /app/docktainer .

# Make the app executable
RUN chmod +x /app/docktainer

# Set up volumes
VOLUME /app/html
VOLUME /app/ssl

# Expose the default ports
EXPOSE 80 443

# Configure the start command
CMD ["sh", "-c", "./docktainer"]