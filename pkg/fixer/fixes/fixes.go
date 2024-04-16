package fixes

type Fix interface {
	Key() string
	Fix(in []byte, opts *RuntimeOptions) (bool, []byte, error)
}

type RuntimeOptions struct {
	Metadata RuntimeMetadata
}

type RuntimeMetadata struct {
	Filename string
}
