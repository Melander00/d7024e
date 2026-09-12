// TODO: Add package documentation for `main`, like this:
// Package main something something...
package main

import (
	"d7024e/internal/kademlia"
	"d7024e/pkg/build"
	"flag"
	"fmt"
)

var (
	BuildVersion string = ""
	BuildTime    string = ""
)

func main() {
	build.BuildVersion = BuildVersion
	build.BuildTime = BuildTime

	mockPtr := flag.Bool("mock", false, "to mock the network")

	flag.Parse()

	if *mockPtr {
		// Start Mockup
		kademlia.MockKademlia(5)
		return
	}

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
