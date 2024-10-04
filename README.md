# go-distributed-storage

## Server Functionality and Features:

- Implements a distributed file system with peer-to-peer communication.
- Manages file storage and retrieval operations across multiple peers.
- Broadcasting: Enables communication with all connected peers.
- Message Handling: Processes messages for file storage and retrieval requests.
- Generic TCP Library Features:

## Transport Interface: Facilitates remote communication between peers.
- Connection Management: Handles dialing and accepting TCP connections.
- Stream Management: Supports sending and closing streams for efficient data transfer.
- Generic Storage Library Features:

- File Management: Provides functionalities for storing, reading, deleting, and checking file existence.
- Path Handling: Supports default root folders and hash-based path generation for unique file storage.
- File Identifier: Encapsulates file path and name for easier manipulation and retrieval.


## Project Features

- **Peer-to-Peer File Storage**: Efficiently store and retrieve files across nodes in the network.
- **TCP-Based Communication**: Utilize TCP for reliable and ordered message delivery.
- **File Management**: Add, remove, and manage files and folders in the storage system.
- **Broadcasting Mechanism**: Support for broadcasting files and messages to connected peers.
- **Custom Decoder**: Implement a custom decoder for managing TCP transport messages.
- **Data Encryption**: Encrypt files for secure storage and transfer.

## Installation

To install and run this project, you need to have Go installed on your system. Then, follow these steps:

1. Clone the repository:
   ```
   git clone https://github.com/AhmedMZaher/go-distributed-storage.git
   ```

2. Change to the project directory:
   ```
   cd go-distributed-storage
   ```

3. Build the project:
   ```
   go build ./cmd/main.go

## API

The `FileServer` struct provides the main functionality:

- `Store(key string, r io.Reader) error`: Stores a file with the given key
- `Get(key string) (io.Reader, error)`: Retrieves a file with the given key
