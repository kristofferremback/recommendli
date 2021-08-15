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
}

// Func is called with the current page's *spotify.Options and PaginatorNextFunc.
// If either paginatorCallbackResult is nil or an error is returned, the pagination is stopped
// and control is returned to the caller.
type Func func(opts *spotify.Options, next NextFunc) (result *NextResult, err error)

type NextFunc func(currentCount, totalCount int) *NextResult

type NextResult struct{ currentCount, totalCount int }

type OptFunc func(p *Paginator)

func New(optFuncs ...OptFunc) *Paginator {
	p := &Paginator{
		pageSize:       50,
		reportProgress: noopProgressReporter,
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

type RunOpts struct {
	offset     int
	totalCount int
	page       int
}

type RunOptsFunc func(r *RunOpts)

func Offset(offset int) RunOptsFunc {
	return func(r *RunOpts) {
		r.offset = offset
	}
}

// Run calls the paginator function until either an error is returned, the *NextResult is nil
// or the current count matches the total count which is set by calling nextFunc(currentCount, totalCount)
// whichever comes first.
// The easiest way to stop the paginator without an error is to `return nil, nil`.
func (p *Paginator) Run(ctx context.Context, paginate Func, runOpts ...RunOptsFunc) error {
	var err error
	if err = ctxhelper.Closed(ctx); err != nil {
		return err
	}

	r := &RunOpts{
		offset:     0,
		totalCount: math.MaxInt64,
		page:       0,
	}
	for _, rOpt := range runOpts {
		rOpt(r)
	}

	for r.offset < r.totalCount {
		if err = ctxhelper.Closed(ctx); err != nil {
			return fmt.Errorf("paginating: %w", err)
		}
		var result *NextResult
		result, err = paginate(&spotify.Options{Limit: &p.pageSize, Offset: &r.offset}, nextFunc)
		if err != nil {
			return fmt.Errorf("paginating: %w", err)
		}
		if result == nil {
			return nil
		}
		r.page++
		r.offset, r.totalCount = result.currentCount, result.totalCount
		p.reportProgress(r.offset, r.totalCount, r.page)
	}
	return nil
}

func nextFunc(currentCount, totalCount int) *NextResult {
	return &NextResult{currentCount: currentCount, totalCount: totalCount}
}
