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
