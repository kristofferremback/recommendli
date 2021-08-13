package spotifypaginator

import (
	"fmt"
	"math"

	"github.com/zmb3/spotify"
)

type ProgressReporterFunc func(currentCount, totalCount, currentPage int)

func noopProgressReporter(currentCount, totalCount, pageCount int) {}

type Paginator struct {
	pageSize       int
	reportProgress ProgressReporterFunc
}

// Func is called with the current page's *spotify.Options and PaginatorNextFunc.
// If either paginatorCallbackResult is nil or an error is returned, the pagination is stopped
// and control is returned to the caller.
type Func func(opts *spotify.Options, next NextFunc) (result *NextResult, err error)

type NextFunc func(currentCount, totalCount int) *NextResult

type NextResult struct{ currentCount, totalCount int }

type OptFuncs func(p *Paginator)

func New(optFuncs ...OptFuncs) *Paginator {
	p := &Paginator{
		pageSize:       50,
		reportProgress: noopProgressReporter,
	}
	for _, optFunc := range optFuncs {
		optFunc(p)
	}
	return p
}

func PageSize(size int) OptFuncs {
	return func(p *Paginator) {
		p.pageSize = size
	}
}

func ProgressReporter(progressReporter ProgressReporterFunc) OptFuncs {
	return func(p *Paginator) {
		p.reportProgress = progressReporter
	}
}

// Run calls the paginator function until either an error is returned, the *NextResult is nil
// or the current count matches the total count which is set by calling nextFunc(currentCount, totalCount)
// whichever comes first.
// The easiest way to stop the paginator without an error is to `return nil, nil`.
func (p *Paginator) Run(paginate Func) error {
	var err error
	currentCount, totalCount := 0, math.MaxInt64
	currentPage := 0
	for currentCount < totalCount {
		var result *NextResult
		result, err = paginate(&spotify.Options{Limit: &p.pageSize, Offset: &currentCount}, nextFunc)
		if err != nil {
			return fmt.Errorf("paginating: %w", err)
		}
		if result == nil {
			return nil
		}
		currentPage++
		currentCount, totalCount = result.currentCount, result.totalCount
		p.reportProgress(currentCount, totalCount, currentPage)
	}
	return nil
}

func nextFunc(currentCount, totalCount int) *NextResult {
	return &NextResult{currentCount: currentCount, totalCount: totalCount}
}
