package spotify

import (
	"fmt"
	"math"

	"github.com/zmb3/spotify"
)

type Paginator struct {
	totalCount int
	limit      int
	offset     int
}

type paginatorFunc func(opts *spotify.Options) (totalCount, offset int, err error)

type paginatorOptFuncs func(p *Paginator)

func newPaginator(optFuncs ...paginatorOptFuncs) *Paginator {
	return &Paginator{
		totalCount: math.MaxInt64,
		limit:      50,
		offset:     0,
	}
}

func (p *Paginator) Options() *spotify.Options {
	return &spotify.Options{Limit: &p.limit, Offset: &p.offset}
}

func (p *Paginator) HasNext(count int) bool {
	return p.totalCount != count
}

func (p *Paginator) Paginate(paginate paginatorFunc) error {
	var err error
	for p.offset < p.totalCount {
		p.totalCount, p.offset, err = paginate(p.Options())
		if err != nil {
			return fmt.Errorf("Error when paginating: %w", err)
		}
	}

	return nil
}

func (p *Paginator) SetTotal(totalCount int) {
	p.totalCount = totalCount
}

func limit(limit int) paginatorOptFuncs {
	return func(p *Paginator) {
		p.limit = limit
	}
}
