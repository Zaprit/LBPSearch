// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

type Slot struct {
	ID               int64
	Name             string
	Description      string
	NpHandle         string
	PublishedIn      string
	Game             int16
	FirstPublished   int64
	LastUpdated      int64
	HeartCount       int64
	Background       string
	Icon             []byte
	RootLevel        []byte
	MissingRootLevel bool
}
