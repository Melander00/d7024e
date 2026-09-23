package cli

import (
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

type dataStorer interface {
	Store(data []byte)
}

type putCmd struct {
	node dataStorer
	name string
}

func CLIPut(node dataStorer) *putCmd {
	return &putCmd{
		node: node,
		name: "put",
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
	if err != nil {
		fmt.Println(err)
		return
	}

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
