# Event-Driven Chat Application

This is an event-driven chat application built with NATS JetStream for data persistence and Protocol Buffers (Protobuf) for efficient communication between services. The project is structured into three main sections:

1. **NATS Server**: Handles message streaming and persistence using JetStream.
2. **Chat Server**: The core of the application, managing chat rooms, user interactions, and message broadcasting.
3. **Client Side**: A console-based application for user interaction and real-time communication.

---

## Project Structure

```
.
├── cmd
│   ├── chatapp
│   │   └── main.go      # Entry point for the client application
│   └── service
│       └── main.go      # Entry point for the chat server
├── docker-compose.yml    # Docker Compose file for running the app
├── Dockerfile.chatapp    # Dockerfile for the client application
├── Dockerfile.service    # Dockerfile for the chat server
├── go.mod                # Go module configuration
├── go.sum                # Go dependencies checksum
├── internal
│   ├── client
│   │   └── client.go     # Client-side logic
│   ├── service
│   │   └── chat_server.go# Chat server logic
│   └── store             # Persistence logic for NATS JetStream
├── LICENSE               # License for the project
├── pkg
│   └── nats
│       └── jetstream.go  # JetStream integration and helper functions
├── proto
│   ├── chat_grpc.pb.go   # Generated gRPC code
│   ├── chat.pb.go        # Generated Protobuf code
│   └── chat.proto        # Protobuf schema
└── README.md             # Project documentation
```

---

## Setup and Instructions

### Prerequisites

Ensure the following tools are installed on your system:
- [Go](https://golang.org/dl/)
- [Docker](https://www.docker.com/products/docker-desktop)
- [Protocol Buffers Compiler (protoc)](https://grpc.io/docs/protoc-installation/)

### Build Protobuf Files

If the generated Protobuf files (`chat.pb.go` and `chat_grpc.pb.go`) do not exist, run the following command:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/chat.proto
```

### Running the Application

1. **Start the Docker Environment**
   
   Bring up the NATS server using Docker Compose:
   ```bash
   docker compose up
   ```

2. **Start the Chat Server**
   
   Open a terminal and run the chat server:
   ```bash
   go run cmd/service/main.go
   ```

3. **Start the Client Application**
   
   Open a new terminal for each user and run the client application:
   ```bash
   go run cmd/chatapp/main.go -user <username>
   ```
   Replace `<username>` with a unique username for each user.

---

## Known Issues

1. **Docker Compose Configuration**:
   - Currently, the Docker Compose file has an issue with resolving the NATS address. As a workaround, run the NATS server via Docker Compose but start the Go services (chat server and client) outside Docker.

---

Enjoy chatting with this event-driven application! Feel free to contribute or report issues to improve the project.
