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

	"github.com/fatih/color"
	"github.com/srivatsa-bot/bt-p2p/files"
	"github.com/srivatsa-bot/bt-p2p/p2p"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:")
		fmt.Println("  bt seed <file>")
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

		fmt.Printf("\n\n%s %s\n", color.GreenString("Seeding file:"), filePath)
		fmt.Printf("%s %s\n", color.GreenString("File ID:"), fileID)
		fmt.Printf("%s %d\n", color.GreenString("Total chunks:"), chunkCount)
		fmt.Printf("%s %s\n", color.GreenString("To download:"), color.YellowString("bt download %s %d output_file", fileID, chunkCount))

		// Handle file requests
		if err := p2p.HandleFileRequest(h, filePath); err != nil {
			log.Fatal("Failed to setup file handler:", err)
		}

		// Announce file
		if err := p2p.AnnounceFile(ctx, kad, fileID); err != nil {
			log.Fatal("Failed to announce file:", err)
		}

		log.Printf("\n%s\n", color.RedString("Seeding... Press Ctrl+C to stop"))
		<-ctx.Done()

	case "download":
		if len(os.Args) != 5 {
			fmt.Println("Usage:")
			fmt.Println("  bt download <file_id> <chunk_count> <output_file>")
			return
		}

		fileID := os.Args[2]
		chunks, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatal("Invalid chunk count:", os.Args[3])
		}
		output := os.Args[4]

		log.Printf("\n\n%s: %s", color.GreenString("[Searching for file]"), fileID)

		peers, err := p2p.FindProviders(ctx, kad, fileID)
		if err != nil {
			log.Fatal("Failed to find providers:", err)
		}

		log.Printf(color.BlueString("Found %d provider(s)"), len(peers))

		// Create output file
		outFile, err := os.Create(output)
		if err != nil {
			log.Fatal("Failed to create output file:", err)
		}
		defer outFile.Close()

		// Pre-allocate file space for better performance
		totalSize := int64(chunks) * 512 * 1024
		if err := outFile.Truncate(totalSize); err != nil {
			log.Printf("Warning: Failed to pre-allocate file space: %v", err)
		}

		log.Printf(color.BlueString("Starting parallel download of %d chunks..."), chunks)
		startTime := time.Now()

		// Create parallel chunk downloader
		downloader := p2p.NewChunkDownloader(h, peers, outFile, chunks)

		// Start parallel download
		if err := downloader.DownloadChunksParallel(ctx); err != nil {
			log.Printf("Download completed with errors: %v", err)

			// Show which chunks failed
			failedChunks := downloader.GetFailedChunks()
			if len(failedChunks) > 0 {
				log.Printf("Failed chunks: %v", failedChunks)
				log.Printf("You may need to retry or find more peers")
			}
		} else {
			duration := time.Since(startTime)
			log.Printf("%s %v!", color.BlueString("[Download completed successfully in:]"), duration)

			// Calculate download speed
			totalMB := float64(totalSize) / (1024 * 1024)
			speedMBps := totalMB / duration.Seconds()
			log.Printf("%s %.2f MB/s", color.BlueString("[Average speed:]"), speedMBps)
		}

	default:
		fmt.Println("Unknown command:", cmd)
		fmt.Println("Available commands: seed, download")
	}
}
