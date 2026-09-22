// TODO: Add package documentation for `main`, like this:
// Package main something something...
package main

import (
	"context"
	"d7024e/internal/cli"
	"d7024e/internal/kademlia"
	"d7024e/internal/kademlia/contact"
	"d7024e/internal/kademlia/rpc"
	"d7024e/pkg/build"
	"d7024e/pkg/network"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var (
	BuildVersion string = ""
	BuildTime    string = ""
)

func main() {
	build.BuildVersion = BuildVersion
	build.BuildTime = BuildTime

	mockPtr := flag.Bool("mock", false, "to mock the network")

	flag.Parse()

	if *mockPtr {
		// Start Mockup
		kademlia.MockKademlia(5000)
		return
	}

	config := kademlia.KademliaConfig{
		Alpha:   3,
		K:       10,
		Timeout: 5 * time.Second,
		Retries: 5,
	}

	ip, err := localIP()

	if err != nil {
		log.Fatal(err)
	}

	receiver := network.NewChannelNetworkReceiver(5)
	net := network.NewUdpNetwork(receiver)
	dataNet := network.NewTCPDataNetwork()
	address := ip.String() + ":8080"
	go net.Listen(":8080")
	id, _ := contact.NewKademliaIDFromAddress(address)
	me := contact.NewContact(id, address)
	rpc := rpc.CreateRpc(receiver, net, me)

	rpc.SetDataNetwork(dataNet)

	if err := rpc.StartDataPlane(); err != nil {
		log.Fatal(err)
	}

	node := kademlia.NewKademliaNode(config, rpc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node.StartBucketRefresh(ctx)

	checkJoinNetwork(node)

	cli.StartCLI(node, address)
}

func localIP() (net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.Name == "lo" {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP
			if ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			return ip, nil
		}
	}

	return nil, fmt.Errorf("no local IPv4 address found")
}

func checkJoinNetwork(node *kademlia.Kademlia) {
	addr, exists := os.LookupEnv("BOOTSTRAP")

	if !exists {
		return
	}

	boot, err := getBootstrapContact(addr)

	if err != nil {
		log.Fatal(err)
	}

	go node.Join(boot)

	fmt.Printf("%s joining boot at %s\n", node.Me.Address, addr)
}

func getBootstrapContact(addr string) (contact.Contact, error) {

	if addr == "" {
		return contact.Contact{}, errors.New("BOOTSTRAP env not set")
	}

	resolved, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return contact.Contact{}, fmt.Errorf("resolve bootstrap %q: %w", addr, err)
	}

	resolvedAddr := resolved.String()

	id, err := contact.NewKademliaIDFromAddress(resolvedAddr)

	if err != nil {
		return contact.Contact{}, err
	}

	return contact.NewContact(id, resolvedAddr), nil
}
