# Packet-loss lookup reliability experiment

## Setup

The experiment uses a simulated Kademlia network with 100 nodes, `Alpha=3`, `K=10`, zero latency, a 100 ms RPC timeout, and three RPC retries. Each packet-loss configuration is run with seeds `1, 2, 3, 4, 5`:

- Packet-loss probabilities: `0`, `0.05`, `0.10`, `0.20`, `0.30`, `0.40`, and `0.50`.
- Each seed creates a fresh simulation and 100 nodes.
- Node addresses use seeded random IPv4 values and deterministic ports. Their addresses determine their Kademlia IDs.
- The boot node stores 100 random 32-byte values. Each value's SHA-256-derived Kademlia ID is the lookup key.
- Every measurement is a remote `LookupData` call from a non-boot node.

The experiment writes a header for every configuration and seed to `data/packet_loss.txt`. It then writes one `lookup_result` record per lookup. Run it from `labs` with:

```bash
go run -tags experiment ./internal/experiments/packetloss
```

Analyze and plot the resulting log with:

```bash
python data/analyze_packet_loss.py data/packet_loss.txt --output data/packet_loss_results
```

The analysis produces per-lookup measurements, a summary CSV, and `success_rate_vs_packet_loss.png`. The summary reports the mean success rate, sample variance, and standard deviation across the five seed-level success rates.

## Measurement meaning

A lookup is successful only when `LookupData` returns without an error and returns exactly the generated value associated with the requested key. This is a useful end-to-end measurement because the value must traverse the simulated network and the Kademlia lookup must discover the boot node. A dropped request or response eventually becomes an RPC timeout; failed candidates are removed and the lookup continues using other candidates. Returning the exact value also prevents a response containing the wrong or missing data from being counted as success.

The topology and values are generated before measurement, and joining is completed before values are stored and lookups begin. Therefore the reported rate measures lookup reliability under packet loss rather than the reliability of network formation or data placement. The fixed node count, protocol settings, number of lookups, and explicit seeds make configurations comparable and runs repeatable.

## Expected result

Success rate should decrease as packet-loss probability increases. A simple independent-packet reference is `1 - p`, but the Kademlia curve should not be expected to match it exactly: a lookup consists of several request/response exchanges, may retry packets, and may try multiple candidate nodes. Retries and redundant candidates can make lookup success better than a single-packet estimate at low and moderate loss. At high loss, many exchanges fail and the lookup can exhaust its candidates or time out, so success should fall sharply toward zero.

The error bars show standard deviation across independent seeds. Compare the measured curve with the `1 - p` reference, and discuss deviations in terms of lookup path length, retry behavior, candidate redundancy, topology quality, and the finite sample size of five seeds and 100 lookups per seed.
