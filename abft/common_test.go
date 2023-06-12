package abft

import (
	"github.com/metatechchain/meta-base/inter/idx"
	"github.com/metatechchain/meta-base/inter/pos"
	"github.com/metatechchain/meta-base/kvdb"
	"github.com/metatechchain/meta-base/kvdb/memorydb"
	"github.com/metatechchain/meta-base/meta"
	"github.com/metatechchain/meta-base/utils/adapters"
	"github.com/metatechchain/meta-base/vecfc"
)

type applyBlockFn func(block *meta.Block) *pos.Validators

// TestMeta extends Meta for tests.
type TestMeta struct {
	*IndexedMeta

	blocks map[idx.Block]*meta.Block

	applyBlock applyBlockFn
}

// FakeMeta creates empty abft with mem store and equal weights of nodes in genesis.
func FakeMeta(nodes []idx.ValidatorID, weights []pos.Weight, mods ...memorydb.Mod) (*TestMeta, *Store, *EventStore) {
	validators := make(pos.ValidatorsBuilder, len(nodes))
	for i, v := range nodes {
		if weights == nil {
			validators[v] = 1
		} else {
			validators[v] = weights[i]
		}
	}

	openEDB := func(epoch idx.Epoch) kvdb.DropableStore {
		return memorydb.New()
	}
	crit := func(err error) {
		panic(err)
	}
	store := NewStore(memorydb.New(), openEDB, crit, LiteStoreConfig())

	err := store.ApplyGenesis(&Genesis{
		Validators: validators.Build(),
		Epoch:      FirstEpoch,
	})
	if err != nil {
		panic(err)
	}

	input := NewEventStore()

	config := LiteConfig()
	lch := NewIndexedMeta(store, input, &adapters.VectorToDagIndexer{vecfc.NewIndex(crit, vecfc.LiteConfig())}, crit, config)

	extended := &TestMeta{
		IndexedMeta: lch,
		blocks:      map[idx.Block]*meta.Block{},
	}

	blockIdx := idx.Block(0)

	err = extended.Bootstrap(meta.ConsensusCallbacks{
		BeginBlock: func(block *meta.Block) meta.BlockCallbacks {
			blockIdx++
			return meta.BlockCallbacks{
				EndBlock: func() (sealEpoch *pos.Validators) {
					// track blocks
					extended.blocks[blockIdx] = block
					if extended.applyBlock != nil {
						return extended.applyBlock(block)
					}
					return nil
				},
			}
		},
	})
	if err != nil {
		panic(err)
	}

	return extended, store, input
}
