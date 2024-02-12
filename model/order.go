package model

type Order int

const (
	OrderAsc Order = iota
	OrderDesc
)

func (o Order) String() string {
	return [...]string{
		"Asc",
		"Desc",
	}[o]
}
