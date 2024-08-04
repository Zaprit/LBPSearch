package ps3

type DataFormat interface {
	fmtID() []byte
	getData() []byte
}

type IndexEntry struct {
	Key   string
	Value any
}
