// currently in sdk use Iterator(start, end []byte) db.Iterator
// while in vmwasm use Iterator(start, end []byte) types.Iterator
// there is a difference in the interface, which should be consistent

package types

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StoreAdapter struct {
	parent sdk.KVStore
}

func NewStoreAdapter(s sdk.KVStore) *StoreAdapter {
	if s == nil {
		panic("store must not be nil")
	}
	return &StoreAdapter{parent: s}
}
func (s StoreAdapter) Get(key []byte) []byte {
	return s.parent.Get(key)
}

func (s StoreAdapter) Set(key, value []byte) {
	s.parent.Set(key, value)
}

func (s StoreAdapter) Delete(key []byte) {
	s.parent.Delete(key)
}

func (s StoreAdapter) Iterator(start, end []byte) wasmvmtypes.Iterator {
	return s.parent.Iterator(start, end)
}

func (s StoreAdapter) ReverseIterator(start, end []byte) wasmvmtypes.Iterator {
	return s.parent.ReverseIterator(start, end)
}
