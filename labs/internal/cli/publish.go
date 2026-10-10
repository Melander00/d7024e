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
	flags := make([]string, 0)

	subargs := make([]string, 0)

	for _, a := range args[1:] {
		b := strings.TrimSpace(a)
		if strings.HasPrefix(b, "--") {
			flags = append(flags, b)
		} else if b != "" {
			subargs = append(subargs, b)
		}
	}

	// fmt.Printf("%d %s", len(subargs), subargs)

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

	fmt.Printf("  publishing %s:%s:%d with size %s\n", domain, pkg, version, formatBytes(data))

	// publish
	cmd.node.Publish(domain, pkg, version, data, force, prev)
}

func (cmd *publishCmd) getName() string {
	return cmd.name
}

func (cmd *publishCmd) getHelp() string {
	return `publish [--force] [--prev=PREVIOUS_VERSION] DOMAIN:PACKAGE:VERSION FILENAME - publishes the package stored in FILENAME`
}
