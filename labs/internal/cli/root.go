package cli

import (
	"bufio"
	"d7024e/internal/kademlia"
	"fmt"
	"os"
	"sort"
	"strings"
)

// This module should handle the CLI part. That means printing and receiving commands.

type CLICommand interface {
	handle([]string)
	getName() string
	getHelp() string
}

type CLI struct {
	node *kademlia.Kademlia
	ip   string
	cmds map[string]CLICommand
}

func StartCLI(node *kademlia.Kademlia, address string) {
	ip := strings.Split(address, ":")[0]

	cli := &CLI{
		node: node,
		ip:   ip,
		cmds: make(map[string]CLICommand),
	}

	cli.registerCommand(CLIExit(node))
	cli.registerCommand(CLIGet(node))
	cli.registerCommand(CLIPing(node))
	cli.registerCommand(CLIPut(node))
	cli.registerCommand(CLIShow(node))

	cli.runCLI()
}

func (cli *CLI) runCLI() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("%s> ", cli.ip)

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		args := strings.Fields(input)

		cmd := args[0]

		if cmd == "help" {
			cli.printHelpString()
			continue
		}

		command, exists := cli.cmds[cmd]

		if exists {
			command.handle(args)
		}
	}
}

func (cli *CLI) registerCommand(cmd CLICommand) {
	name := cmd.getName()
	cli.cmds[name] = cmd
}

func (cli *CLI) printHelpString() {
	fmt.Println("Available commands:")

	names := make([]string, 0, len(cli.cmds))
	for name := range cli.cmds {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		cmd := cli.cmds[name]
		fmt.Printf("  %-15s %s\n", cmd.getName(), cmd.getHelp())
	}
}
