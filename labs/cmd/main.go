// TODO: Add package documentation for `main`, like this:
// Package main something something...
package main

import (
	"d7024e/pkg/build"
	"fmt"
)

var (
	BuildVersion string = ""
	BuildTime    string = ""
)

func main() {
	build.BuildVersion = BuildVersion
	build.BuildTime = BuildTime

	// TODO
	// Start Kademlia
	// Open tty with cli commands.
	fmt.Println("Starting kademlia node...")
}

// NOTE: NewKademliaID is deprecated use ParseKademliaID instead.
// func main() {
// 	fmt.Println("Pretending to run the kademlia app...")
// 	// Using stuff from the kademlia package here. Something like...
// 	id := kademlia.NewKademliaID("FFFFFFFF00000000000000000000000000000000000000000000000000000000")
// 	contact := kademlia.NewContact(id, "localhost:8000")
// 	fmt.Println(contact.String())
// 	fmt.Printf("%v\n", contact)
// }
