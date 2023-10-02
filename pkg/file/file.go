package file

import "errors"

var ErrStorageNotDefined = errors.New("storage not defined")
var ErrNotFound = errors.New("file not found")

type ID string
type StorageType string

const (
	StorageTypeInMemory StorageType = "in_memory"
	StorageTypePostgres StorageType = "postgres"
)

type Link struct {
	StorageType StorageType
	ID          ID
}
