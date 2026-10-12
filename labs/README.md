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


```
$ docker attach kadlab-bootstrap-1

172.25.0.3> publish boot.se:app:1 /app/data/document.txt
	Published boot.se:app:1 with size 86 B
```

```
$ docker run --rm -it \
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
172.25.0.5:8080 joining boot at bootstrap:8080
```

```
172.25.0.3> publish boot.se:app:1 /app/data/scalability.txt --force --prev=0000000000000000000000000000000000000000000000000000000000000000
	Published boot.se:app:1 with size 0.4 MB
172.25.0.3> show ds
Keys stored in local data store
  4d04944b...4ac72a49 => version-record boot.se:app:1 
	key=4d04944bbe1de03d59e403c609384d7949f9c879809a105f7cb247484ac72a49
	sign=/AEX0JsCixWNp0XOM7xpcI9IUd+nOQzU/koHd3/tb0lag143AONOutYN9ZBZRYW2PfnKlP3rTYaBrQ/c7juTCA==
  e9189633...ad808526 => blob
  bb68cbcb...e3f5ce9d => version-record boot.se:app:1 
	key=bb68cbcb91a2969534c2021f2106c92ea8e798504a1e47fb289188a4e3f5ce9d
	sign=VUfU09sYDEmyf09W34CHVl/o247bKhAnQ47r8zKrzjJiXtwIzPdQIs+2dbFwH3CrV+gqqb3CJtdilWvTzMErDw==
  e37012e4...e8acb76f => latest-pointer boot.se:app:1 
	key=e37012e4574945662031baa518d29dbc839d53c66743bbb3f031c8f0e8acb76f
	sign=5r8JKvEQCh++dpCg2bcvIkY4uE7Csr+nidqUwA4SPa1Wj/rpxIJQw/dJu0tWwEPzeIbWA53h6vlsnRoafwALDA==
  17624b6d...3db9f497 => blob
```

```$ docker attach kadlab-node-1

rollback or conflicting version rejected
172.25.0.4> show ds
Keys stored in local data store
  e37012e4...e8acb76f => latest-pointer boot.se:app:1 
	key=e37012e4574945662031baa518d29dbc839d53c66743bbb3f031c8f0e8acb76f
	sign=fLitDik1t+vWntiiv3/Pyk6y3/JU5w6WGrBGhhJhQT9ERDXZhwxdFOpb2Ak2kFIb5164Bl2rnsRU/KDitC1cCw==
  17624b6d...3db9f497 => blob
  4d04944b...4ac72a49 => version-record boot.se:app:1 
	key=4d04944bbe1de03d59e403c609384d7949f9c879809a105f7cb247484ac72a49
	sign=/AEX0JsCixWNp0XOM7xpcI9IUd+nOQzU/koHd3/tb0lag143AONOutYN9ZBZRYW2PfnKlP3rTYaBrQ/c7juTCA==
  e9189633...ad808526 => blob
  bb68cbcb...e3f5ce9d => version-record boot.se:app:1 
	key=bb68cbcb91a2969534c2021f2106c92ea8e798504a1e47fb289188a4e3f5ce9d
	sign=VUfU09sYDEmyf09W34CHVl/o247bKhAnQ47r8zKrzjJiXtwIzPdQIs+2dbFwH3CrV+gqqb3CJtdilWvTzMErDw==
```

```
172.25.0.5> show ds
Keys stored in local data store
  17624b6d...3db9f497 => blob
  4d04944b...4ac72a49 => version-record boot.se:app:1 
	key=4d04944bbe1de03d59e403c609384d7949f9c879809a105f7cb247484ac72a49
	sign=/AEX0JsCixWNp0XOM7xpcI9IUd+nOQzU/koHd3/tb0lag143AONOutYN9ZBZRYW2PfnKlP3rTYaBrQ/c7juTCA==
  e37012e4...e8acb76f => latest-pointer boot.se:app:1 
	key=e37012e4574945662031baa518d29dbc839d53c66743bbb3f031c8f0e8acb76f
	sign=5r8JKvEQCh++dpCg2bcvIkY4uE7Csr+nidqUwA4SPa1Wj/rpxIJQw/dJu0tWwEPzeIbWA53h6vlsnRoafwALDA==
```

Note that node-late (172.25.0.5) has latest-pointer with different signature than node-1 (172.25.0.4) indicating that those are two different records.