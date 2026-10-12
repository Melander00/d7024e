# UDP network

Start the network with docker compose. Change scale node for how many nodes to start.
```bash
docker compose up --build --scale node=2
```

You can attach to the CLI using:
```bash
docker compose attach bootstrap
docker attach kadlab-node-X
```
where X is a specific non-bootstrap node.
Use keyboard shortcuts `Ctrl+P` followed by `Ctrl+Q` to detach.


# Genesis Forks

Start network
```bash
docker compose up --scale node=1
```

Attach to bootstrap and publish original
```bash
docker attach kadlab-bootstrap-1
bootstrap-1 > publish boot.se:app:1 /app/data/document.txt
```

Add new node to the network. It has nothing in its datastore.
```bash
docker run --rm -it \
  --name kadlab-node-late \
  --network kadlab_kademlia \
  -e DNS="http://dns:3000" \
  -e BOOTSTRAP="bootstrap:8080" \
  -e DOMAIN="kadlab-node-late.com" \
  -v "$(pwd)/data:/app/data" \
  -v "$(pwd)/downloads:/app/downloads" \
  --entrypoint sh \
  kadlab-node \
  -c 'export DOMAIN="${HOSTNAME}.com"; exec /app/kademlia-bin'
```

Publish the "malicious" new blob with same package name.
```bash
bootstrap-1 > publish boot.se:app:1 /app/data/scalability.txt --force --prev=0000000000000000000000000000000000000000000000000000000000000000
```

The late node accepts the new latest pointer.
```bash
kadlab-node-late > show ds
```

The node from the start rejected the malicious pointer.
```bash
kadlab-node-1 > show ds
```

We now have two different latest-pointer pointing to the same version with different blobs.

A simple fix in the code would be to remove the `--prev` flag. This, however, is only client-side safety which isn't safety at all.
Enforcing `prev == p.VersionRecordHash` client-sided helps remove accidental genesis forks