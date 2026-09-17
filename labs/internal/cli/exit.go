package cli

import "d7024e/internal/kademlia"

/*

exit
-- terminate the node.

*/

type exitCmd struct {
	node *kademlia.Kademlia
	name string
}

func CLIExit(node *kademlia.Kademlia) *exitCmd {
	return &exitCmd{
		node: node,
		name: "exit",
	}
}

func (cmd *exitCmd) handle(args []string) {

}

func (cmd *exitCmd) getName() string {
	return cmd.name
}

func (cmd *exitCmd) getHelp() string {
	return `exit - stops the node`
}
