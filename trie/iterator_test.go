// Copyright 2014 The go-coupe Authors
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

import "testing"

func TestIterator(t *testing.T) {
	trie := newEmpty()
	vals := []struct{ k, v string }{
		{"do", "verb"},
		{"ether", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"dog", "puppy"},
		{"somethingveryoddindeedthis is", "myothernodedata"},
	}
	v := make(map[string]bool)
	for _, val := range vals {
		v[val.k] = false
		trie.Update([]byte(val.k), []byte(val.v))
	}
	trie.Commit()

	it := NewIterator(trie)
	for it.Next() {
		v[string(it.Key)] = true
	}

	for k, found := range v {
		if !found {
			t.Error("iterator didn't find", k)
		}
	}
}
