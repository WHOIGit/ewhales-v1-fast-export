# Build stage
FROM golang:1.21 AS builder

WORKDIR /app

# Install protoc dependencies
RUN apt-get update && apt-get install -y protobuf-compiler

# Download go modules
COPY go.mod go.sum ./
RUN go mod download

# Install protobuf go generators
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy source code
COPY . .

# Generate protobufs
RUN protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/export.proto

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /ewhales-local ./cmd/local
RUN CGO_ENABLED=0 GOOS=linux go build -o /ewhales-server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /ewhales-client ./cmd/client

# Final stage for server
FROM alpine:latest AS server
WORKDIR /app
COPY --from=builder /ewhales-server /app/ewhales-server
COPY server_config.json /app/
EXPOSE 8443
CMD ["/app/ewhales-server", "-config", "server_config.json"]

# Final stage for client
FROM alpine:latest AS client
WORKDIR /app
COPY --from=builder /ewhales-client /app/ewhales-client
COPY client_config.json /app/
CMD ["/app/ewhales-client", "-config", "client_config.json"]

# Final stage for local
FROM alpine:latest AS local
WORKDIR /app
COPY --from=builder /ewhales-local /app/ewhales-local
COPY config.json /app/
CMD ["/app/ewhales-local", "-config", "config.json"]
