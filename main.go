package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/srivatsa-bot/bt-p2p/files"
	"github.com/srivatsa-bot/bt-p2p/p2p"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  bt seed <file>")
		fmt.Println("  bt download <file_id> <output_file>")
		fmt.Println("  bt download <file_id> <chunk_count> <output_file>")
		return
	}

	// Get first argument(seed or download)
	cmd := os.Args[1]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown when terminated
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel() //cancel the context leading to termination of go routines and other functions
	}()

	// Create host and dht
	h, kad, err := p2p.CreateHost(ctx)
	if err != nil {
		log.Fatal("Failed to create host:", err)
	}
	defer h.Close()

	switch cmd {
	case "seed":
		if len(os.Args) != 3 {
			fmt.Println("Usage: bt seed <file>")
			return
		}

		filePath := os.Args[2]

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Fatal("File does not exist:", filePath)
		}

		// Calculate chunk count
		chunkCount, err := files.ChunkCount(filePath)
		if err != nil {
			log.Fatal("Failed to calculate chunks:", err)
		}

		// Generate file hash as ID
		hash, err := files.FileHash(filePath)
		if err != nil {
			log.Fatal("Failed to hash file:", err)
		}
		fileID := fmt.Sprintf("%x", hash)[:16] // Use first 16 chars of hash

		log.Printf("Seeding file: %s", filePath)
		log.Printf("File ID: %s", fileID)
		log.Printf("Total chunks: %d", chunkCount)
		log.Printf("To download: bt download %s %d <output_file>", fileID, chunkCount)

		// Handle file requests
		if err := p2p.HandleFileRequest(h, filePath); err != nil {
			log.Fatal("Failed to setup file handler:", err)
		}

		// Announce file
		if err := p2p.AnnounceFile(ctx, kad, fileID); err != nil {
			log.Fatal("Failed to announce file:", err)
		}

		log.Println("Seeding... Press Ctrl+C to stop")
		<-ctx.Done()

	case "download":
		if len(os.Args) != 5 {
			fmt.Println("Usage:")
			fmt.Println("  bt download <file_id> <chunk_count> <output_file>")
			return
		}

		fileID := os.Args[2]
		var chunks int
		var output string

		// bt download <file_id> <chunk_count> <output_file>
		var err error
		chunks, err = strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatal("Invalid chunk count:", os.Args[3])
		}
		output = os.Args[4]

		log.Printf("Searching for file: %s", fileID)

		peers, err := p2p.FindProviders(ctx, kad, fileID)
		if err != nil {
			log.Fatal("Failed to find providers:", err)
		}

		log.Printf("Found %d provider(s)", len(peers))

		//create output file
		outFile, err := os.Create(output)
		if err != nil {
			log.Fatal("Failed to create output file:", err)
		}
		defer outFile.Close()

		log.Printf("Downloading %d chunks...", chunks)

		// Download chunks with error handling
		failed := 0
		for i := 0; i < chunks; i++ {
			// Try each peer for this chunk
			downloaded := false
			for _, peer := range peers {
				if err := p2p.RequestChunk(ctx, h, peer, i, outFile); err != nil {
					log.Printf("Failed to download chunk %d from %s: %v", i, peer.ID, err)
					continue
				}
				downloaded = true
				break
			}

			if !downloaded {
				log.Printf("Failed to download chunk %d from any peer", i)
				failed++
			}

			// Add small delay between chunks
			time.Sleep(100 * time.Millisecond)
		}

		if failed > 0 {
			log.Printf("Download completed with %d failed chunks", failed)
		} else {
			log.Println("Download completed successfully!")
		}

	default:
		fmt.Println("Unknown command:", cmd)
		fmt.Println("Available commands: seed, download")
	}
}
