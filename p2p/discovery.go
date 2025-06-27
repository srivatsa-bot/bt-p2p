// this file lets a node announce its seeding a specific file and look for peers who are seeding a specific file. Its a tracker
package p2p

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	peer "github.com/libp2p/go-libp2p/core/peer"
	mh "github.com/multiformats/go-multihash"
)

// Creates a proper content id from a string.
// DTH identifies files in network through this cid
func createCID(s string) (cid.Cid, error) {
	// Hash the string
	hash := sha256.Sum256([]byte(s))

	// Create a multihash(protocol+hash length+hash) from the hash
	mhash, err := mh.EncodeName(hash[:], "sha2-256")
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to create multihash: %w", err)
	}

	// Create CID v1
	c := cid.NewCidV1(cid.Raw, mhash)
	return c, nil
}

func AnnounceFile(ctx context.Context, kad *dht.IpfsDHT, fileID string) error {
	key := "/bt/file/" + fileID

	// Create proper cid for the dht key, values are peer id
	c, err := createCID(key)
	if err != nil {
		return fmt.Errorf("failed to create CID for file %s: %w", fileID, err)
	}

	//provides other peers dth with cid
	err = kad.Provide(ctx, c, true)
	if err != nil {
		return fmt.Errorf("failed to announce file %s: %w", fileID, err)
	}

	log.Printf("\n%s: %s (CID: %s)\n", color.GreenString("Announced file"), key, c.String())
	return nil
}

// finds the peers with the specific cid from dth
func FindProviders(ctx context.Context, kad *dht.IpfsDHT, fileID string) ([]peer.AddrInfo, error) {
	key := "/bt/file/" + fileID

	// Create proper CID for the key
	c, err := createCID(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create CID for file %s: %w", fileID, err)
	}

	log.Printf("%s: %s (CID: %s)\n", color.GreenString("[Searching for providers of]"), key, c.String())

	//context with timeout for provider search
	searchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	//gets peerid and stores them in channel
	provChan := kad.FindProvidersAsync(searchCtx, c, 10)

	var results []peer.AddrInfo
	timeout := time.After(30 * time.Second)

	for {
		select {
		case p, ok := <-provChan:
			if !ok {
				// Channel closed, return results
				if len(results) == 0 {
					return nil, fmt.Errorf("no providers found for file %s", fileID)
				}
				return results, nil
			}
			fmt.Printf("%s %s\n", color.BlueString("Found provider:"), p.ID)
			results = append(results, p)

		case <-timeout:
			if len(results) == 0 {
				return nil, fmt.Errorf("timeout: no providers found for file %s", fileID)
			}
			return results, nil
		}
	}
}
