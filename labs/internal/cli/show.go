package cli

import (
	"d7024e/internal/kademlia"
	"d7024e/internal/kademlia/contact"
	"fmt"
)

/*

show rt
-- print the routing table in an appropriate, human-readable format (useful for debugging and demonstration).

show ds
-- print the data store, i.e. which keys are stored (useful for debugging and demonstration).

*/

type showCmd struct {
	node *kademlia.Kademlia
	name string
}

func CLIShow(node *kademlia.Kademlia) *showCmd {
	return &showCmd{
		node: node,
		name: "show",
	}
}

func (cmd *showCmd) handle(args []string) {
	if len(args) == 1 {
		fmt.Println(cmd.getHelp())
		return
	}

	sub := args[1]

	if sub == "rt" {
		buckets := cmd.node.Routing.GetBuckets()

		if len(buckets) == 0 {
			fmt.Println("Routing table is empty.")
			return
		}

		fmt.Println("Routing Table:")

		for i, bucket := range buckets {
			if bucket.Len() == 0 {
				continue
			}

			contacts := bucket.GetContacts()

			fmt.Printf("  bucket %d:\n", i)
			for _, c := range contacts {
				fmt.Printf("    %s: %s\n", c.Address, truncateId(*c.ID))
			}
		}

	} else if sub == "ds" {
		keys := cmd.node.Datastore.Keys()
		fmt.Println("Keys stored in local data store")
		for _, k := range keys {
			fmt.Printf("  %s\n", truncateId(*k))
		}
	} else {
		fmt.Println(cmd.getHelp())
	}
}

func (cmd *showCmd) getName() string {
	return cmd.name
}

func (cmd *showCmd) getHelp() string {
	return `show <rt|ds> - prints either routing table or keys in data store`
}

func truncateId(id contact.KademliaID) string {
	length := 8

	str := id.String()

	return str[0:length] + "..." + str[len(str)-length:]
}
