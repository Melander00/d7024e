package cli

import (
	"d7024e/internal/kademlia"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type publishCmd struct {
	node *kademlia.Kademlia
	name string
}

func CLIPublish(node *kademlia.Kademlia) *publishCmd {
	return &publishCmd{
		node: node,
		name: "publish",
	}
}

func (cmd *publishCmd) handle(args []string) {
	flags := make([]string, 5)

	subargs := make([]string, 5)

	for _, a := range args {
		if strings.HasPrefix(a, "--") {
			flags = append(flags, a)
		} else {
			subargs = append(subargs, a)
		}
	}

	if len(subargs) != 2 {
		fmt.Println(cmd.getHelp())
		return
	}

	force := false
	prev := ""

	for _, f := range flags {
		if f == "--force" {
			force = true
		} else if strings.HasPrefix(f, "--prev=") {
			splits := strings.Split(f, "=")
			if len(splits) == 2 {
				prev = splits[1]
			}
		}
	}

	packageId := subargs[0]
	filename := subargs[1]

	path := filepath.Join(filename)

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("  publishing %s with size %s\n", filename, formatBytes(data))

	// publish
	cmd.node.Publish(packageId, data, force, prev)
}

func (cmd *publishCmd) getName() string {
	return cmd.name
}

func (cmd *publishCmd) getHelp() string {
	return `publish [--force] [--prev=PREVIOUS_VERSION] DOMAIN:PACKAGE:VERSION FILENAME - publishes the package stored in FILENAME`
}
