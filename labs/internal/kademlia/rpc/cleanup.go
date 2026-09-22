package rpc

import "time"

// For removing old requests, saving memory.

type cleanup struct {
	rpc      *RPC
	maxAlive time.Duration
}

func startGarbageCleaner(rpc *RPC, interval time.Duration, maxAlive time.Duration) {
	gc := cleanup{
		rpc:      rpc,
		maxAlive: maxAlive,
	}

	ticker := time.NewTicker(
		interval,
	)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			gc.clean()
		}
	}
}

func (gc *cleanup) clean() {
	for id, req := range gc.rpc.pending {
		if time.Now().Sub(req.createdAt) >= gc.maxAlive {
			// Delete the request from the map
			gc.rpc.mu.Lock()
			delete(gc.rpc.pending, id)
			gc.rpc.mu.Unlock()
		}
	}
}
