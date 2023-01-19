package etcdsdk

// Op represents an Operation that db can execute
type Op struct {
	// Revision is the revision of a data
	Revision int64
	// ByPrefix enables the prefix scan
	ByPrefix bool
}

type Option func(*Op)

func ByPrefix() Option {
	return func(o *Op) {
		o.ByPrefix = true
	}
}
