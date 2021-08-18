package spotifypaginator

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"golang.org/x/sync/errgroup"
)

type Paginator struct {
	pageSize          int
	initialOffset     int
	initialTotalCount int
	parallelism       int
}

type PageOpts struct {
	Limit  int
	Offset int
}

// Func is called with the current page's PageOpts and PaginatorNextFunc.
// If either paginatorCallbackResult is nil or an error is returned,
// the pagination is stopped and control is returned to the caller.
type Func func(index int, opts PageOpts, next NextFunc) (result *NextResult, err error)

type NextFunc func(totalCount int) *NextResult

type NextResult struct {
	totalCount int
}

type OptFunc func(p *Paginator)

func New(optFuncs ...OptFunc) *Paginator {
	p := &Paginator{
		pageSize:          50,
		initialOffset:     0,
		initialTotalCount: math.MaxInt64,
		parallelism:       3,
	}
	for _, optFunc := range optFuncs {
		optFunc(p)
	}
	return p
}

func Parallelism(parallelism int) OptFunc {
	return func(p *Paginator) {
		p.parallelism = parallelism
	}
}

func PageSize(pageSize int) OptFunc {
	return func(p *Paginator) {
		p.pageSize = pageSize
	}
}

func InitialOffset(offset int) OptFunc {
	return func(p *Paginator) {
		p.initialOffset = offset
	}
}

func InitialTotalCount(totalCount int) OptFunc {
	return func(p *Paginator) {
		p.initialTotalCount = totalCount
	}
}

var errStopPagination = errors.New("stopped")

// RunSync calls the paginator function until either an error is returned, the *NextResult is nil
// or the current count matches the total count which is set by calling nextFunc(offset, totalCount)
// whichever comes first.
// The easiest way to stop the paginator without an error is to `return nil, nil`.
func (p *Paginator) RunSync(ctx context.Context, paginate Func) error {
	return p.run(ctx, paginate, 1)
}

func (p *Paginator) Run(ctx context.Context, paginate Func) error {
	return p.run(ctx, paginate, p.parallelism)
}

func (p *Paginator) run(ctx context.Context, paginate Func, parallelism int) error {
	if err := ctxhelper.Closed(ctx); err != nil {
		return err
	}

	// run initially once to get the total count before we start the parallel iteration
	result, err := paginate(0, p.pageOpts(0, p.initialTotalCount), nextFunc)
	if err != nil {
		return fmt.Errorf("paginating: %w", err)
	}
	if result == nil {
		return nil
	}

	totalCount := result.totalCount
	g, ctx := errgroup.WithContext(ctx)
	guard := make(chan struct{}, parallelism)
	for i := 1; p.pageOpts(i, totalCount).Offset < totalCount; i++ {
		guard <- struct{}{}
		index := i
		g.Go(func() error {
			defer func() {
				<-guard
			}()
			if err := ctxhelper.Closed(ctx); err != nil {
				return err
			}
			result, err := paginate(index, p.pageOpts(index, totalCount), nextFunc)
			if err != nil {
				return err
			}
			if result == nil {
				return errStopPagination
			}
			return nil
		})
	}
	close(guard)
	if err := g.Wait(); err != nil && !errors.Is(err, errStopPagination) {
		return err
	}
	return nil
}

func (p *Paginator) pageOpts(i, max int) PageOpts {
	offset := p.initialOffset + i*p.pageSize
	limit := p.pageSize
	if offset+limit > max {
		limit = max - offset
	}
	return PageOpts{Offset: offset, Limit: limit}
}

func nextFunc(totalCount int) *NextResult {
	return &NextResult{
		totalCount: totalCount,
	}
}
