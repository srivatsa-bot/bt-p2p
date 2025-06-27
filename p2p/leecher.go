package p2p

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
)

// ChunkDownloader manages parallel chunk downloads
type ChunkDownloader struct {
	host        host.Host
	peers       []peer.AddrInfo
	outFile     *os.File
	chunkLocks  []sync.Mutex // Individual mutex for each chunk
	downloaded  []bool       // Track which chunks are downloaded
	failed      []int        // Track failed chunks for retry
	failedMutex sync.Mutex   // Protect failed slice
	totalChunks int
	maxWorkers  int
}

// NewChunkDownloader creates a new parallel chunk downloader
func NewChunkDownloader(h host.Host, peers []peer.AddrInfo, outFile *os.File, totalChunks int) *ChunkDownloader {
	return &ChunkDownloader{
		host:        h,
		peers:       peers,
		outFile:     outFile,
		chunkLocks:  make([]sync.Mutex, totalChunks),
		downloaded:  make([]bool, totalChunks),
		failed:      make([]int, 0),
		totalChunks: totalChunks,
		maxWorkers:  min(10, len(peers)*2), // Limit concurrent workers
	}
}

// DownloadChunksParallel downloads all chunks using goroutines
func (cd *ChunkDownloader) DownloadChunksParallel(ctx context.Context) error {
	// Create job channel for chunk IDs
	jobs := make(chan int, cd.totalChunks)

	// Worker group
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < cd.maxWorkers; i++ {
		wg.Add(1)
		go cd.worker(ctx, jobs, &wg)
	}

	// Send all chunk IDs to jobs channel
	go func() {
		defer close(jobs)
		for i := 0; i < cd.totalChunks; i++ {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for all workers to complete
	wg.Wait()

	// Retry failed chunks
	if len(cd.failed) > 0 {
		log.Printf("Retrying %d failed chunks...", len(cd.failed))
		return cd.retryFailedChunks(ctx)
	}

	return nil
}

// worker processes chunk download jobs
func (cd *ChunkDownloader) worker(ctx context.Context, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case chunkID, ok := <-jobs:
			if !ok {
				return // Channel closed
			}

			// Skip if already downloaded
			if cd.downloaded[chunkID] {
				continue
			}

			// Try to download chunk from available peers
			if err := cd.downloadChunk(ctx, chunkID); err != nil {
				log.Printf("Failed to download chunk %d: %v", chunkID, err)
				cd.addFailedChunk(chunkID)
			}

		case <-ctx.Done():
			return
		}
	}
}

// downloadChunk attempts to download a specific chunk from available peers
func (cd *ChunkDownloader) downloadChunk(ctx context.Context, chunkID int) error {
	// Lock this specific chunk
	cd.chunkLocks[chunkID].Lock()
	defer cd.chunkLocks[chunkID].Unlock()

	// Double-check if chunk was downloaded while waiting for lock
	if cd.downloaded[chunkID] {
		return nil
	}

	// Try each peer until successful
	for _, peerInfo := range cd.peers {
		if err := cd.requestChunkFromPeer(ctx, peerInfo, chunkID); err != nil {
			log.Printf("Failed to download chunk %d from peer %s: %v", chunkID, peerInfo.ID, err)
			continue
		}

		// Mark as downloaded
		cd.downloaded[chunkID] = true
		log.Printf("%s %d", color.GreenString("Successfully downloaded chunk"), chunkID)
		return nil
	}

	return fmt.Errorf(color.RedString("failed to download chunk %d from all peers"), chunkID)
}

// requestChunkFromPeer downloads a chunk from a specific peer
func (cd *ChunkDownloader) requestChunkFromPeer(ctx context.Context, pi peer.AddrInfo, chunkID int) error {
	// Connect to peer with timeout
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := cd.host.Connect(connectCtx, pi); err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", pi.ID, err)
	}

	// Create stream with timeout
	streamCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s, err := cd.host.NewStream(streamCtx, pi.ID, ProtocolID)
	if err != nil {
		return fmt.Errorf("stream creation failed: %w", err)
	}
	defer s.Close()

	// Set deadlines
	s.SetWriteDeadline(time.Now().Add(10 * time.Second))
	s.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Send chunk request
	if _, err := fmt.Fprintf(s, "%d\n", chunkID); err != nil {
		return fmt.Errorf("failed to send chunk request: %w", err)
	}

	// Read response into buffer
	buf := make([]byte, 512*1024)
	n, err := io.ReadFull(s, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("failed to read chunk data: %w", err)
	}

	// Write to file at correct offset
	offset := int64(chunkID) * 512 * 1024
	if _, err := cd.outFile.WriteAt(buf[:n], offset); err != nil {
		return fmt.Errorf("failed to write chunk to file: %w", err)
	}

	return nil
}

// addFailedChunk adds a chunk ID to the failed list
func (cd *ChunkDownloader) addFailedChunk(chunkID int) {
	cd.failedMutex.Lock()
	defer cd.failedMutex.Unlock()
	cd.failed = append(cd.failed, chunkID)
}

// retryFailedChunks retries downloading failed chunks
func (cd *ChunkDownloader) retryFailedChunks(ctx context.Context) error {
	cd.failedMutex.Lock()
	retryList := make([]int, len(cd.failed))
	copy(retryList, cd.failed)
	cd.failed = cd.failed[:0] // Clear failed list
	cd.failedMutex.Unlock()

	// Retry with fewer workers
	retryWorkers := min(3, len(cd.peers))
	jobs := make(chan int, len(retryList))
	var wg sync.WaitGroup

	// Start retry workers
	for i := 0; i < retryWorkers; i++ {
		wg.Add(1)
		go cd.worker(ctx, jobs, &wg)
	}

	// Send failed chunks to retry
	go func() {
		defer close(jobs)
		for _, chunkID := range retryList {
			select {
			case jobs <- chunkID:
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()

	if len(cd.failed) > 0 {
		return fmt.Errorf("still have %d failed chunks after retry", len(cd.failed))
	}

	return nil
}

// GetFailedChunks returns the list of chunks that failed to download
func (cd *ChunkDownloader) GetFailedChunks() []int {
	cd.failedMutex.Lock()
	defer cd.failedMutex.Unlock()
	failed := make([]int, len(cd.failed))
	copy(failed, cd.failed)
	return failed
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
