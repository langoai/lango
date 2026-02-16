package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// Bucket names for the three index orderings.
var (
	bucketSPO = []byte("spo")
	bucketPOS = []byte("pos")
	bucketOSP = []byte("osp")
)

// sep is the null-byte separator between key components.
const sep = '\x00'

// Compile-time check that BoltStore implements Store.
var _ Store = (*BoltStore)(nil)

// BoltStore is a BoltDB-backed triple store with SPO, POS, and OSP indexes.
type BoltStore struct {
	db *bolt.DB
}

// NewBoltStore opens (or creates) a BoltDB database at path and initialises
// the three index buckets.
func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketSPO, bucketPOS, bucketOSP} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return fmt.Errorf("create bucket %s: %w", b, err)
			}
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, fmt.Errorf("init buckets: %w", err)
	}

	return &BoltStore{db: db}, nil
}

// AddTriple adds a single triple to all three indexes.
func (s *BoltStore) AddTriple(_ context.Context, t Triple) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return putTriple(tx, t)
	})
}

// AddTriples adds multiple triples in a single atomic transaction.
func (s *BoltStore) AddTriples(_ context.Context, triples []Triple) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, t := range triples {
			if err := putTriple(tx, t); err != nil {
				return err
			}
		}
		return nil
	})
}

// RemoveTriple removes a triple from all three indexes.
func (s *BoltStore) RemoveTriple(_ context.Context, t Triple) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		spoKey := makeKey(t.Subject, t.Predicate, t.Object)
		posKey := makeKey(t.Predicate, t.Object, t.Subject)
		ospKey := makeKey(t.Object, t.Subject, t.Predicate)

		if err := tx.Bucket(bucketSPO).Delete(spoKey); err != nil {
			return fmt.Errorf("delete spo: %w", err)
		}
		if err := tx.Bucket(bucketPOS).Delete(posKey); err != nil {
			return fmt.Errorf("delete pos: %w", err)
		}
		if err := tx.Bucket(bucketOSP).Delete(ospKey); err != nil {
			return fmt.Errorf("delete osp: %w", err)
		}
		return nil
	})
}

// QueryBySubject returns all triples whose subject matches.
func (s *BoltStore) QueryBySubject(_ context.Context, subject string) ([]Triple, error) {
	prefix := append([]byte(subject), sep)
	var result []Triple

	err := s.db.View(func(tx *bolt.Tx) error {
		triples, err := scanPrefix(tx.Bucket(bucketSPO), prefix, tripleFromSPOKey)
		if err != nil {
			return err
		}
		result = triples
		return nil
	})
	return result, err
}

// QueryByObject returns all triples whose object matches.
func (s *BoltStore) QueryByObject(_ context.Context, object string) ([]Triple, error) {
	prefix := append([]byte(object), sep)
	var result []Triple

	err := s.db.View(func(tx *bolt.Tx) error {
		triples, err := scanPrefix(tx.Bucket(bucketOSP), prefix, tripleFromOSPKey)
		if err != nil {
			return err
		}
		result = triples
		return nil
	})
	return result, err
}

// QueryBySubjectPredicate returns triples matching both subject and predicate.
func (s *BoltStore) QueryBySubjectPredicate(_ context.Context, subject, predicate string) ([]Triple, error) {
	prefix := makeKey(subject, predicate)
	prefix = append(prefix, sep)
	var result []Triple

	err := s.db.View(func(tx *bolt.Tx) error {
		triples, err := scanPrefix(tx.Bucket(bucketSPO), prefix, tripleFromSPOKey)
		if err != nil {
			return err
		}
		result = triples
		return nil
	})
	return result, err
}

// Traverse performs a breadth-first traversal from startNode up to maxDepth
// hops. If predicates is non-empty, only edges with matching predicate types
// are followed.
func (s *BoltStore) Traverse(_ context.Context, startNode string, maxDepth int, predicates []string) ([]Triple, error) {
	predSet := make(map[string]struct{}, len(predicates))
	for _, p := range predicates {
		predSet[p] = struct{}{}
	}

	visited := make(map[string]struct{})
	visited[startNode] = struct{}{}

	frontier := []string{startNode}
	var result []Triple

	err := s.db.View(func(tx *bolt.Tx) error {
		spo := tx.Bucket(bucketSPO)
		osp := tx.Bucket(bucketOSP)

		for depth := 0; depth < maxDepth && len(frontier) > 0; depth++ {
			var nextFrontier []string

			for _, node := range frontier {
				// Outgoing edges: node as subject.
				outgoing, err := scanPrefix(spo, append([]byte(node), sep), tripleFromSPOKey)
				if err != nil {
					return err
				}
				for _, t := range outgoing {
					if len(predSet) > 0 {
						if _, ok := predSet[t.Predicate]; !ok {
							continue
						}
					}
					result = append(result, t)
					if _, seen := visited[t.Object]; !seen {
						visited[t.Object] = struct{}{}
						nextFrontier = append(nextFrontier, t.Object)
					}
				}

				// Incoming edges: node as object.
				incoming, err := scanPrefix(osp, append([]byte(node), sep), tripleFromOSPKey)
				if err != nil {
					return err
				}
				for _, t := range incoming {
					if len(predSet) > 0 {
						if _, ok := predSet[t.Predicate]; !ok {
							continue
						}
					}
					result = append(result, t)
					if _, seen := visited[t.Subject]; !seen {
						visited[t.Subject] = struct{}{}
						nextFrontier = append(nextFrontier, t.Subject)
					}
				}
			}
			frontier = nextFrontier
		}
		return nil
	})
	return result, err
}

// Count returns the total number of triples by counting keys in the SPO bucket.
func (s *BoltStore) Count(_ context.Context) (int, error) {
	var count int
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSPO)
		count = b.Stats().KeyN
		return nil
	})
	return count, err
}

// PredicateStats returns the number of triples grouped by predicate.
func (s *BoltStore) PredicateStats(_ context.Context) (map[string]int, error) {
	stats := make(map[string]int)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSPO)
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			_, p, _, err := splitKey(k)
			if err != nil {
				return err
			}
			stats[p]++
		}
		return nil
	})
	return stats, err
}

// ClearAll removes all triples from all index buckets.
func (s *BoltStore) ClearAll(_ context.Context) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{bucketSPO, bucketPOS, bucketOSP} {
			if err := tx.DeleteBucket(name); err != nil {
				return fmt.Errorf("delete bucket %s: %w", name, err)
			}
			if _, err := tx.CreateBucket(name); err != nil {
				return fmt.Errorf("recreate bucket %s: %w", name, err)
			}
		}
		return nil
	})
}

// Close closes the underlying BoltDB database.
func (s *BoltStore) Close() error {
	return s.db.Close()
}

// --- helpers ---

// makeKey joins components with the null-byte separator.
func makeKey(parts ...string) []byte {
	var buf bytes.Buffer
	for i, p := range parts {
		if i > 0 {
			buf.WriteByte(sep)
		}
		buf.WriteString(p)
	}
	return buf.Bytes()
}

// putTriple writes a triple into all three index buckets within an existing tx.
func putTriple(tx *bolt.Tx, t Triple) error {
	val, err := encodeMetadata(t.Metadata)
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}

	spoKey := makeKey(t.Subject, t.Predicate, t.Object)
	posKey := makeKey(t.Predicate, t.Object, t.Subject)
	ospKey := makeKey(t.Object, t.Subject, t.Predicate)

	if err := tx.Bucket(bucketSPO).Put(spoKey, val); err != nil {
		return fmt.Errorf("put spo: %w", err)
	}
	if err := tx.Bucket(bucketPOS).Put(posKey, val); err != nil {
		return fmt.Errorf("put pos: %w", err)
	}
	if err := tx.Bucket(bucketOSP).Put(ospKey, val); err != nil {
		return fmt.Errorf("put osp: %w", err)
	}
	return nil
}

func encodeMetadata(meta map[string]string) ([]byte, error) {
	if len(meta) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(meta)
}

func decodeMetadata(data []byte) (map[string]string, error) {
	var meta map[string]string
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("decode metadata: %w", err)
	}
	if len(meta) == 0 {
		return nil, nil
	}
	return meta, nil
}

// splitKey splits a null-byte-separated key into exactly 3 components.
func splitKey(key []byte) (string, string, string, error) {
	parts := bytes.SplitN(key, []byte{sep}, 3)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("malformed key: expected 3 parts, got %d", len(parts))
	}
	return string(parts[0]), string(parts[1]), string(parts[2]), nil
}

func tripleFromSPOKey(key, val []byte) (Triple, error) {
	s, p, o, err := splitKey(key)
	if err != nil {
		return Triple{}, err
	}
	meta, err := decodeMetadata(val)
	if err != nil {
		return Triple{}, err
	}
	return Triple{Subject: s, Predicate: p, Object: o, Metadata: meta}, nil
}

func tripleFromOSPKey(key, val []byte) (Triple, error) {
	o, s, p, err := splitKey(key)
	if err != nil {
		return Triple{}, err
	}
	meta, err := decodeMetadata(val)
	if err != nil {
		return Triple{}, err
	}
	return Triple{Subject: s, Predicate: p, Object: o, Metadata: meta}, nil
}

type keyDecoder func(key, val []byte) (Triple, error)

func scanPrefix(b *bolt.Bucket, prefix []byte, decode keyDecoder) ([]Triple, error) {
	var result []Triple
	c := b.Cursor()
	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		t, err := decode(k, v)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}
