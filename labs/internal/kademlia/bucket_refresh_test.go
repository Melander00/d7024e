package kademlia

import (
	"context"
	"testing"
	"time"

	"d7024e/internal/kademlia/contact"
	"d7024e/pkg/network"
)

// TestJoinRefreshesBuckets verifies that Join causes the joining node to learn
// about a node it never contacted directly, via the bootstrap's routing info.
func TestJoinRefreshesBuckets(t *testing.T) {
	simulation := network.NewSimulation()

	bootstrap := createTestNode(t, simulation, "127.0.0.1:11001")
	joining := createTestNode(t, simulation, "127.0.0.1:11002")
	other := createTestNode(t, simulation, "127.0.0.1:11003")

	// Bootstrap already knows about the other node so it can route the joining node toward it.
	bootstrap.Routing.AddContact(other.Me)

	joining.Join(bootstrap.Me)

	found := false
	for _, c := range joining.Routing.FindClosestContacts(other.Me.ID, 10) {
		if c.ID.Equals(other.Me.ID) {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected joining node to discover %s via Join", other.Me.ID.String())
	}
}

// TestRefreshSkipsRecentlyUsedBucket verifies that a bucket which just served a
// lookup is not reported as needing a refresh.
func TestRefreshSkipsRecentlyUsedBucket(t *testing.T) {
	simulation := network.NewSimulation()
	node := createTestNode(t, simulation, "127.0.0.1:11004")

	target := contact.NewRandomKademliaID()
	bucketIndex := node.Routing.BucketIndex(target)

	targetContact := contact.NewContact(target, "")
	node.LookupContact(&targetContact)

	for _, staleIndex := range node.Routing.BucketsNeedingRefresh(time.Hour) {
		if staleIndex == bucketIndex {
			t.Fatalf("expected bucket %d to be excluded after a recent lookup", bucketIndex)
		}
	}
}

// TestStaleBucketNeedsRefresh verifies that a bucket which has never served a
// lookup is reported as needing a refresh.
func TestStaleBucketNeedsRefresh(t *testing.T) {
	simulation := network.NewSimulation()
	node := createTestNode(t, simulation, "127.0.0.1:11005")

	target := contact.NewRandomKademliaID()
	bucketIndex := node.Routing.BucketIndex(target)

	found := false
	for _, staleIndex := range node.Routing.BucketsNeedingRefresh(time.Hour) {
		if staleIndex == bucketIndex {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected unused bucket %d to be reported as stale", bucketIndex)
	}
}

// TestBucketRefreshLoopRuns verifies that the background refresh loop actually
// performs lookups that mark stale buckets as refreshed.
func TestBucketRefreshLoopRuns(t *testing.T) {
	simulation := network.NewSimulation()
	node := createTestNode(t, simulation, "127.0.0.1:11006")
	node.Config.BucketRefreshInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node.StartBucketRefresh(ctx)

	totalBuckets := contact.IDLength * 8
	deadline := time.After(500 * time.Millisecond)
	poll := time.NewTicker(10 * time.Millisecond)
	defer poll.Stop()

	for {
		select {
		case <-poll.C:
			if len(node.Routing.BucketsNeedingRefresh(time.Hour)) < totalBuckets {
				return
			}
		case <-deadline:
			t.Fatal("timed out waiting for the bucket refresh loop to refresh a stale bucket")
		}
	}
}
