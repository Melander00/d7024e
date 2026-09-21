package cli

import (
	"d7024e/internal/kademlia"
	"d7024e/internal/kademlia/contact"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (cmd *putCmd) handle(args []string) {
	if len(args) == 1 {
		fmt.Println(cmd.getHelp())
		return
	}

	target := args[1]

	path := filepath.Join(target)

	data, err := os.ReadFile(path)
	check(err)

	id := contact.NewKademliaIDFromData(data)

	fmt.Printf("  uploading %s with size %s and ID=%s\n", target, formatBytes(data), id.String())

	cmd.node.Store(data)
}

func (cmd *putCmd) getName() string {
	return cmd.name
}

func (cmd *putCmd) getHelp() string {
	return `put <filename> - uploads a file`
}

func formatBytes(b []byte) string {
	size := float64(len(b))

	if size < 1000 {
		return fmt.Sprintf("%d B", len(b))
	}

	units := []string{"kB", "MB", "GB", "TB", "PB"}
	i := int(math.Floor(math.Log(size) / math.Log(1000)))

	return fmt.Sprintf("%.1f %s", size/math.Pow(1000, float64(i+1)), units[i])
}
