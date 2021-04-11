package claimtrie

import (
	"bytes"
	"encoding/gob"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type nodeMgr struct {
	height Height
	db     *leveldb.DB

	cache       map[string]*Node
	nextUpdates todos
}

func newNodeMgr(db *leveldb.DB) *nodeMgr {
	nm := &nodeMgr{
		db:          db,
		cache:       map[string]*Node{},
		nextUpdates: todos{},
	}
	return nm
}

// Load loads the nodes from the database up to height ht.
func (nm *nodeMgr) Load(ht Height) {

	nm.height = ht
	iter := nm.db.NewIterator(nil, nil)
	for iter.Next() {
		name := string(iter.Key())
		nm.cache[name] = nm.load(name, ht)
	}

	data, err := nm.db.Get([]byte("nextUpdates"), nil)
	if err == leveldb.ErrNotFound {
		return
	} else if err != nil {
		panic(err)
	}
	if err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&nm.nextUpdates); err != nil {
		panic(err)
	}
}

// save saves the states to the database.
func (nm *nodeMgr) save() error {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(nm.nextUpdates); err != nil {
		return errors.Wrapf(err, "gob.Encode()")
	}
	if err := nm.db.Put([]byte("nextUpdates"), buf.Bytes(), nil); err != nil {
		return errors.Wrapf(err, "db.Put()")
	}
	return nil
}

// Get returns the latest node with name specified by key.
func (nm *nodeMgr) Get(key []byte) value {
	return nm.nodeAt(string(key), nm.height)
}

// reset resets all nodes to specified height.
func (nm *nodeMgr) reset(ht Height) {
	nm.height = ht
	for name, n := range nm.cache {
		if n.Height >= ht {
			nm.cache[name] = nm.load(name, ht)
		}
	}
}

// Size returns the number of nodes loaded into the cache.
func (nm *nodeMgr) size() int {
	return len(nm.cache)
}

func (nm *nodeMgr) load(name string, ht Height) *Node {
	c := newChangeList(nm.db, name).load().truncate(ht).changes
	return replay(name, c).adjustTo(ht)
}

// nodeAt returns the node adjusted to specified height.
func (nm *nodeMgr) nodeAt(name string, ht Height) *Node {
	n, ok := nm.cache[name]
	if !ok {
		n = NewNode(name)
		nm.cache[name] = n
	}

	// Cached version is too new.
	if n.Height > nm.height || n.Height > ht {
		n = nm.load(name, ht)
	}
	return n.adjustTo(ht)
}

// modifyNode returns the node adjusted to specified height.
func (nm *nodeMgr) modifyNode(name string, chg *change) error {
	ht := nm.height
	n := nm.nodeAt(name, ht)
	n.adjustTo(ht)
	if err := execute(n, chg); err != nil {
		return errors.Wrapf(err, "claim.execute(n,chg)")
	}
	nm.cache[name] = n
	nm.nextUpdates.set(name, ht+1)
	newChangeList(nm.db, name).load().append(chg).save()
	return nil
}

func (nm *nodeMgr) catchUp(ht Height, notifier func(key []byte)) {
	nm.height = ht
	for name := range nm.nextUpdates[ht] {
		notifier([]byte(name))
		if next := nm.nodeAt(name, ht).nextUpdate(); next > ht {
			nm.nextUpdates.set(name, next)
		}
	}
}

// visitFunc visit each node in read-only manner.
type visitFunc func(n *Node) (stop bool)

// Visit visits every node in the cache with VisitFunc.
// If the VisitFunc returns true, the iteration ends immediately.
func (nm *nodeMgr) visit(v visitFunc) {
	for _, n := range nm.cache {
		if v(n) {
			return
		}
	}
}

func replay(name string, chgs []*change) *Node {
	n := NewNode(name)
	for _, chg := range chgs {
		if n.Height < chg.height-1 {
			n.adjustTo(chg.height - 1)
		}
		if n.Height == chg.height-1 {
			if err := execute(n, chg); err != nil {
				panic(err)
			}
		}
	}
	return n
}

func execute(n *Node, c *change) error {
	var err error
	switch c.cmd {
	case cmdAddClaim:
		err = n.addClaim(c.op, c.amount, c.value)
	case cmdSpendClaim:
		err = n.spendClaim(c.op)
	case cmdUpdateClaim:
		err = n.updateClaim(c.op, c.amount, c.id, c.value)
	case cmdAddSupport:
		err = n.addSupport(c.op, c.amount, c.id)
	case cmdSpendSupport:
		err = n.spendSupport(c.op)
	}
	return errors.Wrapf(err, "chg %s", c)
}

type todos map[Height]map[string]bool

func (t todos) set(name string, ht Height) {
	if t[ht] == nil {
		t[ht] = map[string]bool{}
	}
	t[ht][name] = true
}
