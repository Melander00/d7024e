package cli

import "d7024e/internal/kademlia"

/*

put FILENAME
-- upload the contents of a file (a value) under its hash (the key).
Prints the key of the value.

*/

type putCmd struct {
	node *kademlia.Kademlia
	name string
}

func CLIPut(node *kademlia.Kademlia) *putCmd {
	return &putCmd{
		node: node,
		name: "put",
	}
}

func (cmd *putCmd) handle(args []string) {

}

func (cmd *putCmd) getName() string {
	return cmd.name
}

func (cmd *putCmd) getHelp() string {
	return `put <filename> - uploads a file`
}
