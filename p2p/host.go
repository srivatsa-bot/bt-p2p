// this is used to create host or node(seeder/leacher) on the network
package p2p

import (
	"context"
	"fmt"
	"log"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

func CreateHost(ctx context.Context) (host.Host, *dht.IpfsDHT, error) {
	//host(nodes unique identity on network) creation
	h, err := libp2p.New(libp2p.NATPortMap())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}
	//Distributed Hash Table creation(register for nodes)
	kad, err := dht.New(ctx, h, dht.Mode(dht.ModeServer)) //Modeserver option is used so node can store dth reacords nd respond to other peers queries
	if err != nil {
		h.Close()
		return nil, nil, fmt.Errorf("failed to create DHT: %w", err)
	}
	// Connect to bootstrap peers(well maintained nodes that help new peers join into network)
	connected := 0
	for _, addr := range dht.DefaultBootstrapPeers {
		info, err := peer.AddrInfoFromP2pAddr(addr) //fetches peerid from multiaddress
		if err != nil {
			log.Printf("Failed to parse bootstrap peer %s: %v\n", addr, err)
			continue
		}
		//pining bootstrap peers
		if err := h.Connect(ctx, *info); err != nil {
			log.Printf("Bootstrap failed for %s: %v\n", info.ID, err)
		} else {
			log.Printf("Connected to bootstrap peer: %s\n", info.ID)
			connected++
		}
	}

	if connected == 0 {
		log.Println("Warning!!!: No bootstrap peers connected")
	}

	// Bootstrap the DHT
	if err := kad.Bootstrap(ctx); err != nil {
		log.Printf("DHT bootstrap warning: %v\n", err)
	}

	// Wait a bit for DHT to initialize
	time.Sleep(2 * time.Second)

	fmt.Printf("🆔 Peer ID: %s\n", h.ID())
	for _, addr := range h.Addrs() {
		fullAddr := addr.Encapsulate(ma.StringCast("/p2p/" + h.ID().String()))
		fmt.Printf("📡 Address: %s\n", fullAddr)
	}

	return h, kad, nil
}
