package cli

import (
	"d7024e/internal/kademlia"
	"fmt"
	"os"
	"path/filepath"
)

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
	if len(args) == 1 {
		fmt.Println(cmd.getHelp())
		return
	}

	hash := args[1]

	// TODO: Error handling
	data, err := cmd.node.LookupData(hash)

	if err != nil {
		fmt.Println(err)
		return
	}

	if len(args) >= 3 {

		filename := args[2]
		path := filepath.Join(filename)
		os.WriteFile(path, data, 0644)

	} else {
		fmt.Printf("  DATA: %s\n", string(data))
	}
}

func (cmd *getCmd) getName() string {
	return cmd.name
}

func (cmd *getCmd) getHelp() string {
	return `get <key> [filename] - downloads data and optionally saves it to a file`
}
