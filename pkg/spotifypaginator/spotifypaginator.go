package spotifypaginator

import (
	"context"
	"fmt"
	"math"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"github.com/zmb3/spotify"
)

type ProgressReporterFunc func(currentCount, totalCount, currentPage int)

func noopProgressReporter(currentCount, totalCount, pageCount int) {}

type Paginator struct {
	pageSize       int
	reportProgress ProgressReporterFunc
	initialOffset  int
}

// Func is called with the current page's *spotify.Options and PaginatorNextFunc.
// If either paginatorCallbackResult is nil, the Stop() function is called, or an error is returned,
// the pagination is stopped and control is returned to the caller.
type Func func(opts *spotify.Options, next NextFunc) (result *NextResult, err error)

type NextFunc func(totalCount int) *NextResult

type NextResult struct {
	stop       bool
	totalCount int
}

func (n *NextResult) Stop() *NextResult {
	n.stop = true
	return n
}

type OptFunc func(p *Paginator)

func New(optFuncs ...OptFunc) *Paginator {
	p := &Paginator{
		pageSize:       50,
		reportProgress: noopProgressReporter,
		initialOffset:  0,
	}
	for _, optFunc := range optFuncs {
		optFunc(p)
	}
	return p
}

// PageSize sets the paginator's page size.
// `size` should be between 1 <= size <= 50.
func PageSize(size int) OptFunc {
	return func(p *Paginator) {
		p.pageSize = size
	}
}

func ProgressReporter(progressReporter ProgressReporterFunc) OptFunc {
	return func(p *Paginator) {
		p.reportProgress = progressReporter
	}
}

func InitialOffset(offset int) OptFunc {
	return func(p *Paginator) {
		p.initialOffset = offset
	}
}

// Run calls the paginator function until either an error is returned, the *NextResult is nil
// or the current count matches the total count which is set by calling nextFunc(currentCount, totalCount)
// whichever comes first.
// The easiest way to stop the paginator without an error is to `return nil, nil`.
func (p *Paginator) Run(ctx context.Context, paginate Func) error {
	var err error
	if err = ctxhelper.Closed(ctx); err != nil {
		return err
	}

	offset := p.initialOffset
	totalCount := math.MaxInt64
	page := 0

	for offset < totalCount {
		if err = ctxhelper.Closed(ctx); err != nil {
			return fmt.Errorf("paginating: %w", err)
		}
		var result *NextResult
		result, err = paginate(&spotify.Options{Offset: &offset, Limit: &p.pageSize}, nextFunc)
		if err != nil {
			return fmt.Errorf("paginating: %w", err)
		}
		if result == nil {
			return nil
		}

		page++
		offset += p.pageSize
		totalCount = result.totalCount
		p.reportProgress(offset, totalCount, page)
		if result.stop {
			break
		}
	}
	return nil
}

func nextFunc(totalCount int) *NextResult {
	return &NextResult{
		stop:       false,
		totalCount: totalCount,
	}
}
