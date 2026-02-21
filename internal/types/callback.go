package types

// EmbedCallback is an optional hook called when content is saved, enabling
// asynchronous embedding without importing the embedding package.
type EmbedCallback func(id, collection, content string, metadata map[string]string)

// ContentCallback is an optional hook called when content is saved, enabling
// asynchronous processing without importing external packages.
type ContentCallback func(id, collection, content string, metadata map[string]string)

// Triple mirrors graph.Triple to avoid import cycles.
type Triple struct {
	Subject   string
	Predicate string
	Object    string
	Metadata  map[string]string
}

// TripleCallback is an optional hook for saving graph triples.
type TripleCallback func(triples []Triple)
