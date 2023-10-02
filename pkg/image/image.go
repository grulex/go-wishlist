package image

type ID string

type Image struct {
	ID    ID
	Hash  Hash
	Sizes []Size
}

type Size struct {
	Width  uint
	Height uint
	Url    string
}

type Hash struct {
	AHash string
	DHash string
	PHash string
}
