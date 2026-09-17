package cli

import (
	"d7024e/internal/kademlia"
	"d7024e/internal/kademlia/contact"
	"fmt"
	"time"
)

/*

ping IP:PORT
-- ping a node. If you want to be cool, also allow pinging by ID (looking up the IP:port in the routing table, possibly (even cooler) by any unique ID prefix).
Prints the round-trip-time if successful.

*/

type pingCmd struct {
	node *kademlia.Kademlia
	name string
}

func CLIPing(node *kademlia.Kademlia) *pingCmd {
	return &pingCmd{
		node: node,
		name: "ping",
	}
}

func (cmd *pingCmd) handle(args []string) {
	if len(args) == 1 {
		fmt.Println(cmd.getHelp())
		return
	}

	target := args[1]

	fmt.Printf("PING %s\n", target)

	send := time.Now()

	// TODO: maybe abstract away so RPC doesnt need to be called directly

	_, err := cmd.node.Rpc.Ping(contact.Contact{Address: target})
	if err != nil {
		fmt.Println(err)
	}

	rtt := time.Now().Sub(send)
	fmt.Printf("  from %s rtt=%.2f ms\n", target, rtt.Seconds()*1000)
}

func (cmd *pingCmd) getName() string {
	return cmd.name
}

func (cmd *pingCmd) getHelp() string {
	return `ping <ip:port> - pings a node and prints RTT`
}
