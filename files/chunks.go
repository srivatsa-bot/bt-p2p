package files

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

const ChunkSize = 512 * 1024 // 512 KB

// Calculate no of chunks
func ChunkCount(filePath string) (int, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get info about file %s: %w", filePath, err)
	}

	size := info.Size()
	chunks := int(size / ChunkSize)
	if size%ChunkSize != 0 {
		chunks++
	}
	return chunks, nil
}

// reads a specific chunk from the file and returns its data
// used on seeder side
func ReadChunk(file *os.File, chunkID int) ([]byte, error) {

	offset := int64(chunkID) * ChunkSize

	//moves file pointer to start of chunk
	//io.seekstart tells compiler to measure offset from begining.
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to chunk %d: %w", chunkID, err)
	}

	buf := make([]byte, ChunkSize)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read chunk %d: %w", chunkID, err)
	}
	return buf[:n], nil
}

// used on leecher side to download chunks
func WriteChunk(file *os.File, chunkID int, data []byte) error {
	offset := int64(chunkID) * ChunkSize
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to chunk %d: %w", chunkID, err)
	}

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write chunk %d: %w", chunkID, err)
	}
	return nil
}

// Gives sha of 1 chunk
func ChunkHash(data []byte) [32]byte {
	return sha256.Sum256(data)
}

// Returns sha of entire file
func FileHash(filePath string) ([32]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to open file for hashing: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return [32]byte{}, fmt.Errorf("failed to hash file: %w", err)
	}

	var result [32]byte
	copy(result[:], hash.Sum(nil)) //copy hash value from hasher into result
	return result, nil
}
