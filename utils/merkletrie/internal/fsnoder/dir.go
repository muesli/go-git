package fsnoder

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"

	"srcd.works/go-git.v4/utils/merkletrie/noder"
)

// Dir values implement directory-like noders.
type dir struct {
	name     string        // relative
	children []noder.Noder // sorted by name
	hash     []byte        // memoized
}

type byName []noder.Noder

func (a byName) Len() int      { return len(a) }
func (a byName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool {
	return strings.Compare(a[i].Name(), a[j].Name()) < 0
}

// copies the children slice, so nobody can modify the order of its
// elements from the outside.
func newDir(name string, children []noder.Noder) (*dir, error) {
	cloned := make([]noder.Noder, len(children))
	_ = copy(cloned, children)
	sort.Sort(byName(cloned))

	if hasChildrenWithNoName(cloned) {
		return nil, fmt.Errorf("non-root inner nodes cannot have empty names")
	}

	if hasDuplicatedNames(cloned) {
		return nil, fmt.Errorf("children cannot have duplicated names")
	}

	return &dir{
		name:     name,
		children: cloned,
	}, nil
}

func hasChildrenWithNoName(children []noder.Noder) bool {
	for _, c := range children {
		if c.Name() == "" {
			return true
		}
	}

	return false
}

func hasDuplicatedNames(children []noder.Noder) bool {
	if len(children) < 2 {
		return false
	}

	for i := 1; i < len(children); i++ {
		if children[i].Name() == children[i-1].Name() {
			return true
		}
	}

	return false
}

func (d *dir) Hash() []byte {
	if d.hash == nil {
		d.calculateHash()
	}

	return d.hash
}

// hash is calculated as the hash of "dir " plus the concatenation, for
// each child, of its name, a space and its hash.  Children are sorted
// alphabetically before calculating the hash, so the result is unique.
func (d *dir) calculateHash() {
	h := fnv.New64a()
	h.Write([]byte("dir "))
	for _, c := range d.children {
		h.Write([]byte(c.Name()))
		h.Write([]byte(" "))
		h.Write(c.Hash())
	}
	d.hash = h.Sum([]byte{})
}

func (d *dir) Name() string {
	return d.name
}

func (d *dir) IsDir() bool {
	return true
}

// returns a copy so nobody can alter the order of its elements from the
// outside.
func (d *dir) Children() ([]noder.Noder, error) {
	clon := make([]noder.Noder, len(d.children))
	_ = copy(clon, d.children)
	return clon, nil
}

func (d *dir) NumChildren() (int, error) {
	return len(d.children), nil
}

const (
	dirStartMark  = '('
	dirEndMark    = ')'
	dirElementSep = ' '
)

// The string generated by this method is unique for each tree, as the
// children of each node are sorted alphabetically by name when
// generating the string.
func (d *dir) String() string {
	var buf bytes.Buffer

	buf.WriteString(d.name)
	buf.WriteRune(dirStartMark)

	for i, c := range d.children {
		if i != 0 {
			buf.WriteRune(dirElementSep)
		}
		buf.WriteString(c.String())
	}

	buf.WriteRune(dirEndMark)

	return buf.String()
}