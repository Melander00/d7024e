package cli

import "d7024e/internal/kademlia"

/*

get KEY [FILENAME]
-- download the value associated with the key (as a hex string representation of a hash) and save it to the specified filename or print it if no filename is given.
Prints the node it was received from if successful.

*/

type getCmd struct {
	node *kademlia.Kademlia
	name string
}

func CLIGet(node *kademlia.Kademlia) *getCmd {
	return &getCmd{
		node: node,
		name: "get",
	}
}

func (cmd *getCmd) handle(args []string) {

}

func (cmd *getCmd) getName() string {
	return cmd.name
}

func (cmd *getCmd) getHelp() string {
	return `get <key> [filename] - downloads a file and optionally saves the file`
}
