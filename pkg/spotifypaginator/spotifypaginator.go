package spotifypaginator

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"github.com/zmb3/spotify"
	"golang.org/x/sync/errgroup"
)

type ProgressReporterFunc func(offset, totalCount, currentPage int)

func noopProgressReporter(offset, totalCount, pageCount int) {}

type Paginator struct {
	pageSize         int
	reportProgress   ProgressReporterFunc
	initialOffset    int
	concurrencyLimit int
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
		pageSize:         50,
		reportProgress:   noopProgressReporter,
		initialOffset:    0,
		concurrencyLimit: 5,
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

func Concurrency(concurrencyLimit int) OptFunc {
	return func(p *Paginator) {
		p.concurrencyLimit = concurrencyLimit
	}
}

func InitialOffset(offset int) OptFunc {
	return func(p *Paginator) {
		p.initialOffset = offset
	}
}

// Run calls the paginator function until either an error is returned, the *NextResult is nil
// or the current count matches the total count which is set by calling nextFunc(offset, totalCount)
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

type runState struct {
	offset     int
	totalCount int
	page       int
}

type pOpts struct {
	runState runState
	pageSize int
}

func (p pOpts) spotify() *spotify.Options {
	return &spotify.Options{Offset: &p.runState.offset, Limit: &p.pageSize}
}

func (p *Paginator) RunAsync(ctx context.Context, paginate Func) error {
	if err := ctxhelper.Closed(ctx); err != nil {
		return err
	}

	rs := &runState{
		offset:     p.initialOffset,
		totalCount: 0,
		page:       0,
	}

	opts := &spotify.Options{Offset: &rs.offset, Limit: &p.pageSize}
	result, err := paginate(opts, nextFunc)
	if err != nil {
		return fmt.Errorf("paginating: %w", err)
	}
	if result == nil {
		return nil
	}

	rs.page++
	rs.offset += p.pageSize
	rs.totalCount = result.totalCount

	p.reportProgress(rs.offset, rs.totalCount, rs.page)
	if result.stop {
		return nil
	}

	popts := make([]pOpts, 0)
	counter := 0

	for offset := rs.offset; offset < rs.totalCount; offset += p.pageSize {
		popts = append(popts, pOpts{
			runState: runState{offset: offset, totalCount: rs.totalCount, page: rs.page + counter},
			pageSize: p.pageSize,
		})
	}

	errStopped := errors.New("stopped")
	eg, ctx := errgroup.WithContext(ctx)
	guard := make(chan struct{}, p.concurrencyLimit)
	for _, o := range popts {
		guard <- struct{}{}
		spotifyOpts, rs := o.spotify(), o.runState
		eg.Go(func() error {
			defer func() {
				<-guard
			}()
			if err := ctxhelper.Closed(ctx); err != nil {
				return err
			}
			result, err := paginate(spotifyOpts, nextFunc)
			if err != nil {
				return err
			}
			if result == nil || result.stop {
				return errStopped
			}
			p.reportProgress(rs.offset, rs.totalCount, rs.page)
			return nil
		})
	}
	close(guard)
	if err := eg.Wait(); err != nil && !errors.Is(err, errStopped) {
		return err
	}

	return nil

	// optsChan := make(chan pOpts, p.concurrencyLimit)

	// errStopped := errors.New("stopped")
	// g, ctx := errgroup.WithContext(ctx)
	// g.Go(func() error {
	// 	for opts := range optsChan {
	// 		if err := ctxhelper.Closed(ctx); err != nil {
	// 			return err
	// 		}
	// 		spotifyOpts, rs := opts.spotify(), opts.runState
	// 		result, err := paginate(spotifyOpts, nextFunc)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		if result == nil || result.stop {
	// 			return errStopped
	// 		}
	// 		p.reportProgress(rs.offset, rs.totalCount, rs.page)
	// 	}
	// 	return nil
	// })

	// counter2 := 0
	// for offset := rs.offset; offset < rs.totalCount; offset += p.pageSize {
	// 	optsChan <- pOpts{
	// 		runState: runState{offset: offset, totalCount: rs.totalCount, page: rs.page + counter2},
	// 		pageSize: p.pageSize,
	// 	}
	// }
	// close(optsChan)

	// err = g.Wait()
	// if err != nil && !errors.Is(err, errStopped) {
	// 	return err
	// }

	// return nil
}

func nextFunc(totalCount int) *NextResult {
	return &NextResult{
		stop:       false,
		totalCount: totalCount,
	}
}
