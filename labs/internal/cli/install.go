package cli

import (
	"d7024e/internal/kademlia"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	packageId := args[1]
	splits := strings.Split(packageId, ":")
	if len(splits) != 3 {
		return
	}

	domain := splits[0]
	pkg := splits[1]
	version := splits[2]

	if cmd.node == nil {
		fmt.Println("node is not initialized")
		return
	}
	blobHash, err := cmd.findBlobHash(domain, pkg, version)

	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	data, err := cmd.node.LookupData(blobHash)

	if err != nil {
		fmt.Println(err)
		return
	}

	filename := packageId
	path := filepath.Join("/app/downloads", filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Println(err)
		return
	}
}

func (cmd *installCmd) findBlobHash(domain string, pkg string, version string) (string, error) {
	latest, err := cmd.node.GetLatestVersion(domain, pkg)
	if err != nil {
		return "", err
	}

	recHash := latest.VersionRecordHash
	rec, err := cmd.node.GetVersionRecord(recHash)
	if err != nil {
		return "", err
	}

	if version == "latest" {
		return rec.BlobHash, nil
	}

	vers, err := strconv.ParseUint(version, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid version, should be latest or a uint64")
	}

	cont := true
	found := false
	for cont {
		// find package with lastPackageHash
		rec, err = cmd.node.GetVersionRecord(recHash)
		if err != nil {
			return "", err
		}

		if rec.Version == vers {
			found = true
			break
		}

		if rec.PreviousVersionRecord != "" && rec.PreviousVersionRecord != zeroHash() {
			recHash = rec.PreviousVersionRecord
		} else {
			cont = false
		}
	}

	if found {
		return rec.BlobHash, nil
	} else {
		return "", fmt.Errorf("version not found")
	}

}

func (cmd *installCmd) getName() string {
	return cmd.name
}

func (cmd *installCmd) getHelp() string {
	return `publish [--force] [--prev=PREVIOUS_VERSION] DOMAIN:PACKAGE:VERSION FILENAME - publishes the package stored in FILENAME`
}
