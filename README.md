# go-distributed-storage

- A decentralized file storage system implemented in Go, allowing for peer-to-peer file storage and retrieval across a network of nodes. 
- The transport library used in this project is generic and can be utilized for any application requiring network communication.

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
