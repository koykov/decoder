package decoder

import (
	"sync"

	"github.com/koykov/byteconv"
)

// Decoders database.
// Contains two indexes describes two types of pairs between decoders and keys/IDs.
type db struct {
	mux sync.RWMutex
	// ID index. Value is an offset in the dec array.
	idxID map[int]int
	// Key index. Value is an offset in the dec array as well.
	idxKey map[string]int
	// Hash index. Value is an offset in the dec array.
	idxHash map[uint64]int
	// Decoders storage.
	buf []*Decoder
}

func initDB() *db {
	return &db{
		idxID:   make(map[int]int),
		idxKey:  make(map[string]int),
		idxHash: make(map[uint64]int),
	}
}

// Save decoder tree in the storage and make two pairs (ID-dec and key-dec).
func (db *db) set(id int, key string, tree *Tree) {
	dec := Decoder{
		ID:   id,
		Key:  key,
		tree: tree,
	}
	db.mux.Lock()
	var idx int
	if idx = db.getIdxLF(id, key); idx >= 0 && idx < len(db.buf) {
		db.buf[idx] = &dec
	} else {
		db.buf = append(db.buf, &dec)
		idx = len(db.buf) - 1
		if id >= 0 {
			db.idxID[id] = idx
		}
		if key != "-1" {
			db.idxKey[key] = idx
		}
	}
	if _, ok := db.idxHash[tree.hsum]; !ok {
		db.idxHash[tree.hsum] = idx
	}
	db.mux.Unlock()
}

// Get first decoder found by key or ID.
func (db *db) get(id int, key string) (dec *Decoder) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if idx := db.getIdxLF(id, key); idx >= 0 && idx < len(db.buf) {
		dec = db.buf[idx]
	}
	return
}

// Lock-free index getter.
//
// Returns first available index by key or ID.
func (db *db) getIdxLF(id int, key string) (idx int) {
	idx = -1
	if idx1, ok := db.idxKey[key]; ok && idx1 != -1 {
		idx = idx1
	} else if idx1, ok := db.idxID[id]; ok && idx1 != -1 {
		idx = idx1
	}
	return
}

// Get decoder by ID.
func (db *db) getID(id int) *Decoder {
	return db.get(id, "-1")
}

// Get decoder by key.
func (db *db) getKey(key string) *Decoder {
	return db.get(-1, key)
}

// Get decoder by key and fallback key.
func (db *db) getKey1(key, key1 string) (dec *Decoder) {
	idx := -1
	db.mux.RLock()
	defer db.mux.RUnlock()
	idx1, ok := db.idxKey[key]
	if !ok {
		idx1, ok = db.idxKey[key1]
	}
	if ok {
		idx = idx1
	}
	if idx >= 0 && idx < len(db.buf) {
		dec = db.buf[idx]
	}
	return
}

// Get parsed tree by hash sum.
func (db *db) getTreeByHash(hsum uint64) *Tree {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if idx, ok := db.idxHash[hsum]; ok && idx >= 0 && idx < len(db.buf) {
		return db.buf[idx].tree
	}
	return nil
}

// Get decoder by list of keys describes as bytes arrays.
func (db *db) getBKeys(bkeys [][]byte) (dec *Decoder) {
	l := len(bkeys)
	if l == 0 {
		return
	}
	db.mux.RLock()
	defer db.mux.RUnlock()
	_ = bkeys[l-1]
	for i := 0; i < l; i++ {
		if idx, ok := db.idxKey[byteconv.B2S(bkeys[i])]; ok && idx >= 0 && idx < len(db.buf) {
			dec = db.buf[idx]
			return
		}
	}
	return
}
