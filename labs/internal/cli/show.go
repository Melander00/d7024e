package cli

import "d7024e/internal/kademlia"

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

}

func (cmd *showCmd) getName() string {
	return cmd.name
}

func (cmd *showCmd) getHelp() string {
	return `show <rt|ds> - prints either routing table or keys in data store`
}
