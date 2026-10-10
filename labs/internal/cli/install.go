package cli

import (
	"d7024e/internal/kademlia"
	"fmt"
)

type installCmd struct {
	node *kademlia.Kademlia
	name string
}

func CLIInstall(node *kademlia.Kademlia) *installCmd {
	return &installCmd{
		node: node,
		name: "install",
	}
}

func (cmd *installCmd) handle(args []string) {
	if len(args) != 2 {
		fmt.Println(cmd.getHelp())
		return
	}
	// pick out version
	// packageId := args[1]
	// splits := strings.Split(packageId, ":")
	// if splits.len != 3 reutrn
	// version := splits[2]
}

func (cmd *installCmd) getName() string {
	return cmd.name
}

func (cmd *installCmd) getHelp() string {
	return `publish [--force] [--prev=PREVIOUS_VERSION] DOMAIN:PACKAGE:VERSION FILENAME - publishes the package stored in FILENAME`
}
