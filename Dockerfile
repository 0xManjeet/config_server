FROM golang:latest

# Create a directory for persistent storage
RUN mkdir -p /data

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 51051

# Define a volume for persistent storage
VOLUME ["/data"]

ENTRYPOINT ["/entrypoint.sh"]