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

type DetailedNews struct {
	OneNews     News
	AllComments []Comments
}

type Pagination struct {
	NumOfPages int
	Page       int
	Limit      int
}

type ResponseWithPagiantion struct {
	AllNews []News
	Pg      Pagination
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
