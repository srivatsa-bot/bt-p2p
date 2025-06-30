# P2P File Discovery Client

A high-performance peer-to-peer file discovery and sharing client built on [libp2p](https://libp2p.io/) and IPFS DHT (Distributed Hash Table). This client enables decentralized file discovery across networks with automatic peer connectivity and relay support.

## 🚀 Features

- **Distributed File Announcement**: Announce files on the DHT network for discovery by other peers
- **Intelligent Peer Discovery**: Multi-stage provider search with aggressive fallback mechanisms
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
go build -o p2p-client .
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

## 🎯 Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "your-project/p2p"
    dht "github.com/libp2p/go-libp2p-kad-dht"
)

func main() {
    ctx := context.Background()
    
    // Initialize your libp2p host and DHT (implementation details omitted)
    host, kad := setupP2P() // Your setup function
    
    // Announce a file
    fileID := "my-important-file-123"
    err := p2p.AnnounceFile(ctx, kad, fileID)
    if err != nil {
        log.Fatalf("Failed to announce file: %v", err)
    }
    
    // Find providers for a file
    providers, err := p2p.FindProviders(ctx, kad, fileID)
    if err != nil {
        log.Fatalf("Failed to find providers: %v", err)
    }
    
    // Ensure connectivity to found providers
    connectedPeers := p2p.EnsurePeerConnectivity(host, kad, providers)
    log.Printf("Connected to %d providers", len(connectedPeers))
}
```

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

**Example:**
```go
err := p2p.AnnounceFile(ctx, kad, "document-abc123")
if err != nil {
    log.Printf("Announcement failed: %v", err)
}
```

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

**Example:**
```go
providers, err := p2p.FindProviders(ctx, kad, "document-abc123")
if err != nil {
    log.Printf("No providers found: %v", err)
    return
}

for _, provider := range providers {
    fmt.Printf("Found provider: %s with %d addresses\n", 
        provider.ID, len(provider.Addrs))
}
```

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

```go
// Provider search timeout
searchCtx, cancel := context.WithTimeout(ctx, 60*time.Second)

// Maximum providers to find
provChan := kad.FindProvidersAsync(searchCtx, c, 20)

// Re-announcement interval
ticker := time.NewTicker(10 * time.Minute)

// Connection timeout
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
```

### Network Configuration

For optimal performance across different network topologies:

1. **Public Networks**: Default settings work well
2. **Private Networks**: Consider reducing timeouts and increasing re-announcement frequency
3. **High-Latency Networks**: Increase search and connection timeouts

## 🚨 Error Handling

The client includes comprehensive error handling:

- **Network Timeouts**: Graceful handling of slow network conditions
- **Provider Discovery Failures**: Automatic fallback to aggressive search
- **Connection Failures**: Relay path discovery for unreachable peers
- **Re-announcement Failures**: Logged but non-blocking for continuous operation

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

**Aggressive Search Triggered**
```
No providers found, trying aggressive search...
```
- Normal behavior when initial search fails
- Indicates sparse network or few providers
- Consider increasing search timeout for better results

## 🔍 Monitoring and Logging

The client provides detailed logging with color-coded output:

- **Green**: Successful operations (announcements, discoveries)
- **Blue**: Provider information
- **Standard**: Informational messages
- **Error logs**: Connection failures and timeouts

Enable debug logging for detailed DHT operations:
```go
log.SetLevel(log.DebugLevel)
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [libp2p](https://libp2p.io/) - Modular network stack
- [IPFS](https://ipfs.io/) - DHT implementation
- [Go](https://golang.org/) - Programming language

## 📞 Support

- Create an issue for bug reports or feature requests
- Check existing issues before creating new ones
- Provide detailed reproduction steps for bugs

---

**Built with ❤️ using libp2p and Go**
