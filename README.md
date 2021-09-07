# go-multistore

This repository provides a mechanism for constructing multiple, isolated, IPFS storage instances (blockstore, filestore, DAGService) on top of a single
go-datastore instance.

### Background reading

You may want to familiarize yourself with various IPFS storage layer components:

- [DataStore](https://github.com/ipfs/go-datastore)
- [BlockStore](https://github.com/ipfs/go-ipfs-blockstore)
- [FileStore](https://github.com/ipfs/go-filestore)
- [BlockService](https://github.com/ipfs/go-blockservice)
- [DAGService](https://github.com/ipfs/go-ipld-format/blob/master/merkledag.go)

## Installation
```bash
go get "github.com/filecoin-project/go-multistore"`
```

## Usage

Initialize multistore:

```golang
var ds datastore.Batching
multiDs, err := multistore.NewMultiDstore(ds)
```

Create new store:

```golang
next := multiDs.Next()
store, err := multiDs.Get(store)

// store will have a blockstore, filestore, and DAGService
```

List existing store indexes:

```golang
indexes := multiDs.List()
```

Delete a store (will delete all data in isolated store without touching the rest of the datastore):

```golang
var index int
err := multiDs.Delete(index)
```

Shutdown (make sure everything is closed):

```golang
multiDs.Close()
```

