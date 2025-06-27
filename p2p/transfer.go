// transfer logic for p2p
package p2p

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const ProtocolID = protocol.ID("/bt/file/1.0.0")

// Runs on seeder side, listens for incomming requests using the mentioned protocol
func HandleFileRequest(h host.Host, filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	h.SetStreamHandler(ProtocolID, func(s network.Stream) {

		defer s.Close()
		log.Printf("Incoming stream from: %s", s.Conn().RemotePeer())

		// Read deadline for seeding
		s.SetReadDeadline(time.Now().Add(30 * time.Second))

		reader := bufio.NewReader(s)
		chunkReq, err := reader.ReadString('\n') //read till newline to get chunkid send by the leecher
		if err != nil {
			log.Printf("Failed to read chunk request: %v", err)
			return
		}
		chunkReq = strings.TrimSpace(chunkReq) //remove newline

		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("File open error: %v", err)
			return
		}
		defer file.Close()

		chunkID, err := strconv.Atoi(chunkReq)
		if err != nil {
			log.Printf("Invalid chunk ID: %s", chunkReq)
			return
		}

		//moves file pointer based on chunk id
		offset := int64(chunkID) * 512 * 1024
		if _, err := file.Seek(offset, 0); err != nil {
			log.Printf("Seek error: %v", err)
			return
		}

		//read chunk into buffer
		buf := make([]byte, 512*1024)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			log.Printf("Read error: %v", err)
			return
		}

		// Set write deadline
		s.SetWriteDeadline(time.Now().Add(30 * time.Second))

		//In last chunk we may send garbage insted of acutal data, so we write until then actual length of data in 512kb buffer
		if _, err := s.Write(buf[:n]); err != nil {
			log.Printf("Write error: %v", err)
			return
		}

		log.Printf("Sent chunk %d (%d bytes)", chunkID, n)
	})

	return nil
}

// Runs on leacher side, requests specific chunk from seeder
func RequestChunk(ctx context.Context, h host.Host, pi peer.AddrInfo, chunkID int, outFile *os.File) error {
	// Connect to peer first
	if err := h.Connect(ctx, pi); err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", pi.ID, err)
	}

	// Create stream with timeout
	streamCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s, err := h.NewStream(streamCtx, pi.ID, ProtocolID)
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

	// Move pointer to offset
	offset := int64(chunkID) * 512 * 1024
	if _, err := outFile.Seek(offset, 0); err != nil {
		return fmt.Errorf("failed to seek in output file: %w", err)
	}

	// Write the bytes from offset
	if _, err := outFile.Write(buf[:n]); err != nil {
		return fmt.Errorf("failed to write chunk to file: %w", err)
	}

	log.Printf("Received chunk %d (%d bytes)", chunkID, n)
	return nil
}
