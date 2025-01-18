package storage

type News struct {
	Id        int
	Title     string
	Content   string
	CreatedAt int64
}

type Comments struct {
	Id        int
	NewsId    int
	Content   string
	CreatedAt int64
}

const (
	FullMatchText = iota
	PartialMatchText
	FullMatchHeader
	PartialMatchHeader
	SelectionDate
	DateRange
	ExcludedPhrases
	FieldSort
)

const (
	DateSort = iota
	NameSort
)
