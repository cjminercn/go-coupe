// Copyright 2015 The go-coupe Authors
// This file is part of the go-coupe library.
//
// The go-coupe library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-coupe library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-coupe library. If not, see <http://www.gnu.org/licenses/>.

package trie

import (
	"bytes"
	"testing"

	"github.com/cjminercn/go-coupe/common"
	"github.com/cjminercn/go-coupe/ethdb"
)

// makeTestTrie create a sample test trie to test node-wise reconstruction.
func makeTestTrie() (ethdb.Database, *Trie, map[string][]byte) {
	// Create an empty trie
	db, _ := ethdb.NewMemDatabase()
	trie, _ := New(common.Hash{}, db)

	// Fill it with some arbitrary data
	content := make(map[string][]byte)
	for i := byte(0); i < 255; i++ {
		key, val := common.LeftPadBytes([]byte{1, i}, 32), []byte{i}
		content[string(key)] = val
		trie.Update(key, val)

		key, val = common.LeftPadBytes([]byte{2, i}, 32), []byte{i}
		content[string(key)] = val
		trie.Update(key, val)
	}
	trie.Commit()

	// Return the generated trie
	return db, trie, content
}

// checkTrieContents cross references a reconstructed trie with an expected data
// content map.
func checkTrieContents(t *testing.T, db Database, root []byte, content map[string][]byte) {
	trie, err := New(common.BytesToHash(root), db)
	if err != nil {
		t.Fatalf("failed to create trie at %x: %v", root, err)
	}
	for key, val := range content {
		if have := trie.Get([]byte(key)); bytes.Compare(have, val) != 0 {
			t.Errorf("entry %x: content mismatch: have %x, want %x", key, have, val)
		}
	}
}

// Tests that an empty trie is not scheduled for syncing.
func TestEmptyTrieSync(t *testing.T) {
	emptyA, _ := New(common.Hash{}, nil)
	emptyB, _ := New(emptyRoot, nil)

	for i, trie := range []*Trie{emptyA, emptyB} {
		db, _ := ethdb.NewMemDatabase()
		if req := NewTrieSync(common.BytesToHash(trie.Root()), db, nil).Missing(1); len(req) != 0 {
			t.Errorf("test %d: content requested for empty trie: %v", i, req)
		}
	}
}

// Tests that given a root hash, a trie can sync iteratively on a single thread,
// requesting retrieval tasks and returning all of them in one go.
func TestIterativeTrieSyncIndividual(t *testing.T) { testIterativeTrieSync(t, 1) }
func TestIterativeTrieSyncBatched(t *testing.T)    { testIterativeTrieSync(t, 100) }

func testIterativeTrieSync(t *testing.T, batch int) {
	// Create a random trie to copy
	srcDb, srcTrie, srcData := makeTestTrie()

	// Create a destination trie and sync with the scheduler
	dstDb, _ := ethdb.NewMemDatabase()
	sched := NewTrieSync(common.BytesToHash(srcTrie.Root()), dstDb, nil)

	queue := append([]common.Hash{}, sched.Missing(batch)...)
	for len(queue) > 0 {
		results := make([]SyncResult, len(queue))
		for i, hash := range queue {
			data, err := srcDb.Get(hash.Bytes())
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x: %v", hash, err)
			}
			results[i] = SyncResult{hash, data}
		}
		if index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		queue = append(queue[:0], sched.Missing(batch)...)
	}
	// Cross check that the two tries re in sync
	checkTrieContents(t, dstDb, srcTrie.Root(), srcData)
}

// Tests that the trie scheduler can correctly reconstruct the state even if only
// partial results are returned, and the others sent only later.
func TestIterativeDelayedTrieSync(t *testing.T) {
	// Create a random trie to copy
	srcDb, srcTrie, srcData := makeTestTrie()

	// Create a destination trie and sync with the scheduler
	dstDb, _ := ethdb.NewMemDatabase()
	sched := NewTrieSync(common.BytesToHash(srcTrie.Root()), dstDb, nil)

	queue := append([]common.Hash{}, sched.Missing(10000)...)
	for len(queue) > 0 {
		// Sync only half of the scheduled nodes
		results := make([]SyncResult, len(queue)/2+1)
		for i, hash := range queue[:len(results)] {
			data, err := srcDb.Get(hash.Bytes())
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x: %v", hash, err)
			}
			results[i] = SyncResult{hash, data}
		}
		if index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		queue = append(queue[len(results):], sched.Missing(10000)...)
	}
	// Cross check that the two tries re in sync
	checkTrieContents(t, dstDb, srcTrie.Root(), srcData)
}

// Tests that given a root hash, a trie can sync iteratively on a single thread,
// requesting retrieval tasks and returning all of them in one go, however in a
// random order.
func TestIterativeRandomTrieSyncIndividual(t *testing.T) { testIterativeRandomTrieSync(t, 1) }
func TestIterativeRandomTrieSyncBatched(t *testing.T)    { testIterativeRandomTrieSync(t, 100) }

func testIterativeRandomTrieSync(t *testing.T, batch int) {
	// Create a random trie to copy
	srcDb, srcTrie, srcData := makeTestTrie()

	// Create a destination trie and sync with the scheduler
	dstDb, _ := ethdb.NewMemDatabase()
	sched := NewTrieSync(common.BytesToHash(srcTrie.Root()), dstDb, nil)

	queue := make(map[common.Hash]struct{})
	for _, hash := range sched.Missing(batch) {
		queue[hash] = struct{}{}
	}
	for len(queue) > 0 {
		// Fetch all the queued nodes in a random order
		results := make([]SyncResult, 0, len(queue))
		for hash, _ := range queue {
			data, err := srcDb.Get(hash.Bytes())
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x: %v", hash, err)
			}
			results = append(results, SyncResult{hash, data})
		}
		// Feed the retrieved results back and queue new tasks
		if index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		queue = make(map[common.Hash]struct{})
		for _, hash := range sched.Missing(batch) {
			queue[hash] = struct{}{}
		}
	}
	// Cross check that the two tries re in sync
	checkTrieContents(t, dstDb, srcTrie.Root(), srcData)
}

// Tests that the trie scheduler can correctly reconstruct the state even if only
// partial results are returned (Even those randomly), others sent only later.
func TestIterativeRandomDelayedTrieSync(t *testing.T) {
	// Create a random trie to copy
	srcDb, srcTrie, srcData := makeTestTrie()

	// Create a destination trie and sync with the scheduler
	dstDb, _ := ethdb.NewMemDatabase()
	sched := NewTrieSync(common.BytesToHash(srcTrie.Root()), dstDb, nil)

	queue := make(map[common.Hash]struct{})
	for _, hash := range sched.Missing(10000) {
		queue[hash] = struct{}{}
	}
	for len(queue) > 0 {
		// Sync only half of the scheduled nodes, even those in random order
		results := make([]SyncResult, 0, len(queue)/2+1)
		for hash, _ := range queue {
			data, err := srcDb.Get(hash.Bytes())
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x: %v", hash, err)
			}
			results = append(results, SyncResult{hash, data})

			if len(results) >= cap(results) {
				break
			}
		}
		// Feed the retrieved results back and queue new tasks
		if index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		for _, result := range results {
			delete(queue, result.Hash)
		}
		for _, hash := range sched.Missing(10000) {
			queue[hash] = struct{}{}
		}
	}
	// Cross check that the two tries re in sync
	checkTrieContents(t, dstDb, srcTrie.Root(), srcData)
}

// Tests that a trie sync will not request nodes multiple times, even if they
// have such references.
func TestDuplicateAvoidanceTrieSync(t *testing.T) {
	// Create a random trie to copy
	srcDb, srcTrie, srcData := makeTestTrie()

	// Create a destination trie and sync with the scheduler
	dstDb, _ := ethdb.NewMemDatabase()
	sched := NewTrieSync(common.BytesToHash(srcTrie.Root()), dstDb, nil)

	queue := append([]common.Hash{}, sched.Missing(0)...)
	requested := make(map[common.Hash]struct{})

	for len(queue) > 0 {
		results := make([]SyncResult, len(queue))
		for i, hash := range queue {
			data, err := srcDb.Get(hash.Bytes())
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x: %v", hash, err)
			}
			if _, ok := requested[hash]; ok {
				t.Errorf("hash %x already requested once", hash)
			}
			requested[hash] = struct{}{}

			results[i] = SyncResult{hash, data}
		}
		if index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		queue = append(queue[:0], sched.Missing(0)...)
	}
	// Cross check that the two tries re in sync
	checkTrieContents(t, dstDb, srcTrie.Root(), srcData)
}
