# P2P File Discovery Client

A high-performance peer-to-peer file discovery and sharing client built on [libp2p](https://libp2p.io/) and IPFS DHT (Distributed Hash Table). This client enables decentralized file discovery across networks with automatic peer connectivity and relay support.

## 🚀 Features

- **Distributed File Announcement**: Announce files on the DHT network for discovery by other peers
- **Intelligent Peer Discovery**: Multi-stage provider search with aggressive fallback mechanisms
- **Parallel Chunk Download**: Downloads files in 512KB chunks using Go routines for maximum speed
- **Memory Safe Transfers**: Mutex locks ensure thread-safe chunk assembly and memory protection
- **Automatic Re-announcement**: Periodic file re-announcement to maintain network presence
- **Relay Support**: Automatic relay path discovery for NAT traversal
- **Cross-Network Discovery**: Support for discovery across different network topologies
- **Connection Management**: Smart peer connectivity with fallback relay mechanisms
- **Colored Logging**: Enhanced logging with color-coded output for better debugging

## 📋 Prerequisites

- Go 1.19 or higher
- Git

## 🛠️ Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/p2p-file-client.git
cd p2p-file-client

# Install dependencies
go mod tidy

# Build the client
go build -o bt .
```

## 📦 Dependencies

This project uses the following key dependencies:

```go
github.com/libp2p/go-libp2p-kad-dht    // DHT implementation
github.com/libp2p/go-libp2p           // Core libp2p functionality  
github.com/ipfs/go-cid                // Content addressing
github.com/multiformats/go-multihash  // Multi-hash support
github.com/fatih/color                // Colored terminal output
```

## 🎯 Usage

The client provides two main commands for file sharing and downloading:

### Seed a File

Share a file on the P2P network, making it discoverable by other peers:

```bash
bt seed <file>
```

**Example:**
```bash
bt seed document.pdf
bt seed /path/to/video.mp4
bt seed large-dataset.zip
```

### Download a File

Download a file from the P2P network using parallel chunk downloading:

```bash
bt download <file_id> <chunk_count> <output_file>
```

**Example:**
```bash
bt download abc123def456 10 downloaded-document.pdf
bt download xyz789uvw012 50 large-video.mp4
bt download def456ghi789 100 massive-dataset.zip
```

**Parameters:**
- `file_id`: Unique identifier of the file to download
- `chunk_count`: Number of 512KB chunks to download in parallel
- `output_file`: Local filename for the downloaded file

## ⚡ Parallel Download Architecture

The client implements high-performance parallel downloading with the following features:

### Concurrent Chunk Processing
- **512KB Chunks**: Files are split into 512KB chunks for optimal network transfer
- **Go Routines**: Each chunk is downloaded concurrently using separate Go routines
- **Configurable Parallelism**: Specify the number of parallel chunks via `chunk_count` parameter

### Memory Safety
- **Mutex Locks**: Thread-safe chunk assembly using mutex synchronization
- **Memory Protection**: Each chunk write operation is protected against race conditions
- **Safe Reassembly**: Chunks are safely combined in correct order to reconstruct the original file

### Performance Benefits
- **Faster Downloads**: Parallel transfers significantly reduce download time
- **Better Utilization**: Maximizes bandwidth usage across multiple peer connections
- **Fault Tolerance**: If one chunk fails, others continue downloading independently
- **Load Distribution**: Downloads spread across multiple providers for better performance

## 📚 API Documentation

### File Announcement

#### `AnnounceFile(ctx context.Context, kad *dht.IpfsDHT, fileID string) error`

Announces a file on the DHT network, making it discoverable by other peers.

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `kad`: IPFS DHT instance
- `fileID`: Unique identifier for the file

**Features:**
- Creates a unique CID (Content Identifier) for the file
- Announces with extended TTL for better availability
- Automatic re-announcement every 10 minutes
- Graceful shutdown on context cancellation

### Provider Discovery

#### `FindProviders(ctx context.Context, kad *dht.IpfsDHT, fileID string) ([]peer.AddrInfo, error)`

Discovers peers that have announced the specified file.

**Parameters:**
- `ctx`: Context for cancellation and timeouts
- `kad`: IPFS DHT instance  
- `fileID`: File identifier to search for

**Returns:**
- Slice of `peer.AddrInfo` containing provider information
- Error if no providers found or search fails

**Features:**
- 60-second search timeout
- Finds up to 20 providers
- Automatic address resolution for peers without addresses
- Fallback to aggressive search if no providers found initially

### Connection Management

#### `EnsurePeerConnectivity(h host.Host, kad *dht.IpfsDHT, peers []peer.AddrInfo) []peer.AddrInfo`

Establishes connections to discovered providers with relay fallback.

**Parameters:**
- `h`: libp2p host instance
- `kad`: IPFS DHT instance
- `peers`: List of peer information to connect to

**Returns:**
- Filtered list of successfully connected peers

**Features:**
- 15-second connection timeout per peer
- Automatic relay path discovery for unreachable peers
- Circuit relay support for NAT traversal

## 🔧 Configuration

### Search Parameters

You can modify search behavior by adjusting these parameters in the code:

- **Provider search timeout**: 60 seconds default
- **Maximum providers to find**: 20 providers default
- **Re-announcement interval**: 10 minutes default
- **Connection timeout**: 15 seconds per peer default

### Network Configuration

For optimal performance across different network topologies:

1. **Public Networks**: Default settings work well
2. **Private Networks**: Consider reducing timeouts and increasing re-announcement frequency
3. **High-Latency Networks**: Increase search and connection timeouts

### Chunk Download Optimization

- **Small Files (<5MB)**: Use 1-5 chunks for minimal overhead
- **Medium Files (5-100MB)**: Use 10-20 chunks for balanced performance
- **Large Files (>100MB)**: Use 50+ chunks for maximum parallelism
- **Network Quality**: Reduce chunk count on slow/unstable connections

## 🚨 Error Handling

The client includes comprehensive error handling:

- **Network Timeouts**: Graceful handling of slow network conditions
- **Provider Discovery Failures**: Automatic fallback to aggressive search
- **Connection Failures**: Relay path discovery for unreachable peers
- **Re-announcement Failures**: Logged but non-blocking for continuous operation
- **Chunk Download Failures**: Failed chunks are retried automatically
- **Memory Safety**: Mutex locks prevent data corruption during parallel operations

## 🐛 Troubleshooting

### Common Issues

**No Providers Found**
```
Error: no providers found for file xyz
```
- Ensure the file was properly announced
- Check network connectivity
- Verify DHT bootstrap nodes are reachable

**Connection Failures**
```
Failed to connect to provider: connection refused
```
- Providers may be behind NAT/firewall
- Relay discovery will attempt automatic workaround
- Ensure relay nodes are available in the network

**Slow Downloads**
```
Download taking too long
```
- Increase chunk count for more parallelism
- Check network bandwidth and latency
- Verify multiple providers are available

**Memory Issues**
```
Out of memory during download
```
- Reduce chunk count to lower memory usage
- Ensure sufficient RAM for parallel operations
- Monitor system resources during large downloads

## 🙏 Acknowledgments

- [libp2p](https://libp2p.io/) - Modular network stack
- [IPFS](https://ipfs.io/) - DHT implementation
- [Go](https://golang.org/) - Programming language and concurrency primitives

## 📞 Support

- Create an issue for bug reports or feature requests
- Check existing issues before creating new ones
- Provide detailed reproduction steps for bugs

---


