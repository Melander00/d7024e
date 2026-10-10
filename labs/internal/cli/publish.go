package cli

import (
	"d7024e/internal/kademlia"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

	splits := strings.Split(subargs[0], ":")
	if len(splits) != 3 {
		fmt.Println("packageid needs to follow DOMAIN:PACKAGE:VERSION")
		return
	}

	domain := splits[0]
	pkg := splits[1]
	version, err := strconv.ParseUint(splits[2], 10, 64)
	if err != nil {
		fmt.Println("version needs to be a uint64")
	}

	filename := subargs[1]

	path := filepath.Join(filename)

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("  publishing %s with size %s\n", filename, formatBytes(data))

	// publish
	cmd.node.Publish(domain, pkg, version, data, force, prev)
}

func (cmd *publishCmd) getName() string {
	return cmd.name
}

func (cmd *publishCmd) getHelp() string {
	return `publish [--force] [--prev=PREVIOUS_VERSION] DOMAIN:PACKAGE:VERSION FILENAME - publishes the package stored in FILENAME`
}
