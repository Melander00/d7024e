package cli

import (
	"d7024e/internal/kademlia"
	"d7024e/internal/kademlia/contact"
	"encoding/base64"
	"fmt"
	"strings"
)

/*

show rt
-- print the routing table in an appropriate, human-readable format (useful for debugging and demonstration).

show ds
-- print the data store, i.e. which keys are stored (useful for debugging and demonstration).

*/

type showCmd struct {
	node *kademlia.Kademlia
	name string
}

func CLIShow(node *kademlia.Kademlia) *showCmd {
	return &showCmd{
		node: node,
		name: "show",
	}
}

func (cmd *showCmd) handle(args []string) {
	if len(args) == 1 {
		fmt.Println(cmd.getHelp())
		return
	}

	sub := args[1]

	if sub == "rt" {
		buckets := cmd.node.Routing.GetBuckets()

		if len(buckets) == 0 {
			fmt.Println("Routing table is empty.")
			return
		}

		fmt.Println("Routing Table:")

		for i, bucket := range buckets {
			if bucket.Len() == 0 {
				continue
			}

			contacts := bucket.GetContacts()

			fmt.Printf("  bucket %d:\n", i)
			for _, c := range contacts {
				fmt.Printf("    %s: %s\n", c.Address, truncateId(*c.ID))
			}
		}

	} else if sub == "ds" {
		keys := cmd.node.Datastore.Keys()
		fmt.Println("Keys stored in local data store")
		for _, k := range keys {
			fmt.Printf("  %s\n", truncateId(*k))
		}

	} else if sub == "dns" {
		if len(args) == 2 {
			fmt.Println(cmd.getHelp())
			return
		}
		domain := args[2]
		pk, err := cmd.node.DNS.LookupPK(domain)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("\t%s => %s\n", domain, base64.URLEncoding.EncodeToString(pk))
	} else {

		splits := strings.Split(sub, ":")
		if len(splits) == 2 {

			domain := splits[0]
			pkg := splits[1]

			latest, err := cmd.node.GetLatestVersion(domain, pkg)
			if err != nil {
				fmt.Printf("error %s\n", err.Error())
				return
			}

			fmt.Printf("Version chain of %s:%s:\n", domain, pkg)

			recHash := latest.VersionRecordHash

			cont := true
			for cont {
				// find package with lastPackageHash
				rec, err := cmd.node.GetVersionRecord(recHash)
				if err != nil {
					fmt.Printf("%s\n", err.Error())
					return
				}

				fmt.Printf("\t%s => %s\n", rec.Version, truncateId(*rec.Hash()))

				if rec.PreviousVersionRecord != "" {
					recHash = rec.PreviousVersionRecord
				} else {
					cont = false
				}
			}

			return
		}

		fmt.Println(cmd.getHelp())
	}
}

func (cmd *showCmd) getName() string {
	return cmd.name
}

func (cmd *showCmd) getHelp() string {
	return `show <rt|ds|dns|DOMAIN:PACKAGE> [dns->domain] - prints either routing table or keys in data store`
}

func truncateId(id contact.KademliaID) string {
	length := 8

	str := id.String()

	return str[0:length] + "..." + str[len(str)-length:]
}
