package errgroupx

import (
	"context"
	"sync"
)

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid and does not cancel on error.
type Group struct {
	ctx context.Context
	err      error
	ch       chan func() error
	errOnce  sync.Once
	cancel   func()
	wg       sync.WaitGroup
}

// WithContext create a Group.
// given function from Go will receive this context,
func WithContext(ctx context.Context, parallel int) (*Group, context.Context) {
	if parallel <= 0 {
		panic("errgroup: parallel must great than 0")
	}
	groupCtx, cancel := context.WithCancel(ctx)
	g := new(Group)
	g.ctx = groupCtx
	g.cancel = cancel
	g.ch = make(chan func() error)
	for i := 0; i < parallel; i++ {
		go func() {
			for f := range g.ch {
				g.do(f)
			}
		}()
	}
	return g, ctx
}

func (g *Group) do(f func() error) {
	var err error
	defer func() {
		if err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
		g.wg.Done()
	}()
	err = f()
}

func (g *Group) Wait() error {
	g.wg.Wait()

	if g.ch != nil {
		close(g.ch) // let all receiver exit
	}

	if g.cancel != nil {
		g.cancel()
	}

	return g.err
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	select {
	case g.ch <- f:
	case <-g.ctx.Done():
		g.wg.Done()
	}
}
