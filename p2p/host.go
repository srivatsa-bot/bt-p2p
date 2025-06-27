// this is used to create host or node(seeder/leacher) on the network
package p2p

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	libp2p "github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
)

const ProtocolID = protocol.ID("/bt/file/1.0.0")

func CreateHost(ctx context.Context) (host.Host, *dht.IpfsDHT, error) {

	//host(nodes unique identity on network) creation
	h, err := libp2p.New(libp2p.NATPortMap())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}
	//Distributed Hash Table creation(register for nodes) Kademila in this instance
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
			log.Printf("%s %s: %v\n", color.RedString("Bootstrap failed for"), info.ID, err)
		} else {
			log.Printf("%s %s\n", color.GreenString("Connected to bootstrap peer:"), info.ID)
			connected++
		}
	}

	if connected == 0 {
		log.Println("Warning!!!: No bootstrap peers connected")
	}

	// Bootstrap the DHT
	//Once connected to bootstrap peers, node will ask info about the peers boottrap connected to and fills it dht
	if err := kad.Bootstrap(ctx); err != nil {
		log.Printf("DHT bootstrap warning: %v\n", err)
	}

	// Wait a bit for DHT to initialize and populate with discovered peers
	time.Sleep(2 * time.Second)
	fmt.Printf("%s", color.BlueString("\n[Client Node Info]\n"))
	fmt.Printf("%s %s\n", color.GreenString("[Peer ID]:"), h.ID())
	//loop through list of ip of this node and adds your Peer ID to each address to form a multiaddress.
	// This address will be recorded in other peers dht's. eg /ipv4/port/p2p/peerid
	for _, addr := range h.Addrs() {
		fullAddr := addr.Encapsulate(ma.StringCast("/p2p/" + h.ID().String()))
		fmt.Printf("%s %s\n", color.GreenString("[Address]:"), fullAddr)
	}

	return h, kad, nil
}
