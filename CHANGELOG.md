# Changelog

## [v1.1.1] - 2024-10-11
### Added
- **New Integration Test for Distributed File System**: 
  - Introduced `TestComplexDFSScenario` to simulate various real-world scenarios, ensuring system reliability and performance. 
  - The test verifies proper startup and inter-node communication among five DFS nodes, performs file operations, simulates node failures, and tests data persistence and recovery.
  - To run the test, use: 
    ```bash
    go test -v -run TestComplexDFSScenario
    ```
- **File Size Limitation Handling**: 
  - Implemented a mechanism to manage file transfers exceeding 100MB, preventing system stalls during file retrieval operations.

## [v1.1.0] - 2024-10-09
### Fixed
- **Multi-Hop File Retrieval Issue**: 
  - Addressed communication and retrieval problems in the peer-to-peer network when the bootstrap node deletes a file.
  
### Changes
- **Introduction of Message Types**:
  - `NodeIntroductionMessage`: Sent by new peers upon joining, contains the peer's address for record updates.
  - `PeersInfoMessage`: Sent by the bootstrap node to provide a complete list of peer addresses to new peers.

- **Logic Implementation**:
  - The `OnPeer` method checks for bootstrap nodes and sends relevant peer information upon new peer connections.
  - Enhanced broadcasting logic for both message types.

### Tests
- **TestMultiHopFileRequest**:
  - Simulates a multi-hop file request within the P2P network, validating communication and retrieval among nodes.


## [v1.0.0] - 2024-10-06
### Added
- Initial release of the project with core functionality.
- File storage system with distributed peer-to-peer communication.
- Basic encryption and decryption support using AES in CTR mode.
- API endpoints for storing and retrieving files.
- Multi-peer server setup with bootstrapping capabilities.
- Tests for storing and retrieving multiple files to ensure reliability.

### Notes
- This is the first stable release. Future updates will include improvements and additional features.
