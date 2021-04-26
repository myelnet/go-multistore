package multistore_test

import (
	"math/rand"
	"testing"

	"github.com/filecoin-project/go-multistore"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	dss "github.com/ipfs/go-datastore/sync"
	"github.com/stretchr/testify/require"
)

func TestMultistore(t *testing.T) {

	ds := dss.MutexWrap(datastore.NewMapDatastore())
	multiDS, err := multistore.NewMultiDstore(ds)
	require.NoError(t, err)

	var stores []*multistore.Store
	for i := 0; i < 5; i++ {
		next := multiDS.Next()
		store, err := multiDS.Get(next)
		require.NoError(t, err)
		stores = append(stores, store)
		blks := generateBlocksOfSize(5, 100)
		for _, block := range blks {
			err := store.Bstore.Put(block)
			require.NoError(t, err)
		}
	}

	t.Run("creates all keys", func(t *testing.T) {
		qres, err := ds.Query(query.Query{KeysOnly: true})
		require.NoError(t, err)
		all, err := qres.Rest()
		require.NoError(t, err)
		require.Len(t, all, 26)
	})

	t.Run("lists stores", func(t *testing.T) {
		storeIndexes := multiDS.List()
		require.Len(t, storeIndexes, 5)
		require.Equal(t, multistore.StoreIDList{1, 2, 3, 4, 5}, storeIndexes)

		// getting a second time does not make a new store
		_, err := multiDS.Get(3)
		require.NoError(t, err)
		storeIndexes = multiDS.List()
		require.Len(t, storeIndexes, 5)
		require.Equal(t, multistore.StoreIDList{1, 2, 3, 4, 5}, storeIndexes)
	})

	t.Run("delete stores", func(t *testing.T) {
		multiDS.Delete(4)
		storeIndexes := multiDS.List()
		require.Len(t, storeIndexes, 4)
		require.Equal(t, multistore.StoreIDList{1, 2, 3, 5}, storeIndexes)

		qres, err := ds.Query(query.Query{KeysOnly: true})
		require.NoError(t, err)
		all, err := qres.Rest()
		require.NoError(t, err)
		require.Len(t, all, 21)
	})

	t.Run("close/reopen", func(t *testing.T) {
		err := multiDS.Close()
		require.NoError(t, err)
		newMultiDS, err := multistore.NewMultiDstore(ds)

		storeIndexes := newMultiDS.List()
		require.Len(t, storeIndexes, 4)
		require.Equal(t, multistore.StoreIDList{1, 2, 3, 5}, storeIndexes)

		next := newMultiDS.Next()
		require.Equal(t, multistore.StoreID(6), next)
	})
}

func TestAddAndDeleteSameDataDifferentStores(t *testing.T) {
	ds := dss.MutexWrap(datastore.NewMapDatastore())
	multiDS, err := multistore.NewMultiDstore(ds)
	require.NoError(t, err)

	blks := generateBlocksOfSize(1, 100)
	targetBlock := blks[0]

	// Create two multi-stores with the same block
	var stores []*multistore.Store
	var storeIDs []multistore.StoreID
	for i := 0; i < 2; i++ {
		next := multiDS.Next()
		store, err := multiDS.Get(next)
		require.NoError(t, err)
		stores = append(stores, store)
		storeIDs = append(storeIDs, next)
		err = store.Bstore.Put(targetBlock)
		require.NoError(t, err)
	}

	// Expect the block to be in both multi-store's blockstore
	_, err = stores[0].Bstore.Get(targetBlock.Cid())
	require.NoError(t, err)
	_, err = stores[1].Bstore.Get(targetBlock.Cid())
	require.NoError(t, err)

	// Delete the first multi-store
	err = multiDS.Delete(storeIDs[0])
	require.NoError(t, err)

	// Expect the block to only be in the second multi-store's blockstore
	_, err = stores[0].Bstore.Get(targetBlock.Cid())
	require.Error(t, err) // expect error because block was deleted from first multi-store
	_, err = stores[1].Bstore.Get(targetBlock.Cid())
	require.NoError(t, err)

	// Delete the other multi-store
	err = multiDS.Delete(storeIDs[1])
	require.NoError(t, err)

	// Expect not to be able to get the block any more from either
	// multi-store's blockstore
	_, err = stores[0].Bstore.Get(targetBlock.Cid())
	require.Error(t, err) // expect error because block is not found
	_, err = stores[1].Bstore.Get(targetBlock.Cid())
	require.Error(t, err) // expect error because block is not found
}

var seedSeq int64 = 0

func randomBytes(n int64) []byte {
	randBytes := make([]byte, n)
	r := rand.New(rand.NewSource(seedSeq))
	_, _ = r.Read(randBytes)
	seedSeq++
	return randBytes
}

func generateBlocksOfSize(n int, size int64) []blocks.Block {
	generatedBlocks := make([]blocks.Block, 0, n)
	for i := 0; i < n; i++ {
		b := blocks.NewBlock(randomBytes(size))
		generatedBlocks = append(generatedBlocks, b)

	}
	return generatedBlocks
}
