package errgroupx

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParallelNoError(t *testing.T) {
	ctx := context.Background()
	g, _ := WithContext(ctx, 10)
	var num int64
	add := func() error {
		atomic.AddInt64(&num, 1)
		return nil
	}
	for i := 0; i < 10; i++ {
		g.Go(add)
	}
	require.NoError(t, g.Wait())
	require.Equal(t, int64(10), num)
}

func TestParallel(t *testing.T) {
	ctx := context.Background()
	g, _ := WithContext(ctx, 10)
	var num int64
	add := func() error {
		atomic.AddInt64(&num, 1)
		return nil
	}
	for i := 0; i < 100; i++ {
		g.Go(add)
	}
	require.NoError(t, g.Wait())
	require.Equal(t, int64(100), num)
}

func TestParallelError(t *testing.T) {
	ctx := context.Background()
	g, _ := WithContext(ctx, 10)
	var num int64
	add := func() error {
		atomic.AddInt64(&num, 1)
		if atomic.CompareAndSwapInt64(&num, 5, 0) {
			return errors.New("reach to 5")
		}
		return nil
	}
	for i := 0; i < 100; i++ {
		g.Go(add)
	}

	require.Error(t, g.Wait())
	// require.Equal(t, int64(0), num)
}

func TestParallelLimit(t *testing.T) {
	ctx := context.Background()
	g, _ := WithContext(ctx, 2)
	sleep := func() error {
		fmt.Println("sleep 1 second")
		time.Sleep(time.Second)
		return nil
	}
	now := time.Now()
	for i := 0; i < 10; i++ {
		g.Go(sleep)
	}

	require.NoError(t, g.Wait())
	require.WithinDuration(t, now, time.Now(), 6*time.Second)
}

func TestParallelLimitError(t *testing.T) {
	ctx := context.Background()
	g, _ := WithContext(ctx, 2)
	var num int64 = 1
	sleep := func() error {
		atomic.AddInt64(&num, 1)
		if atomic.CompareAndSwapInt64(&num, 3, 0) {
			return errors.New("reach to 5")
		}
		fmt.Println("sleep", num)
		time.Sleep(time.Second)
		return nil
	}
	now := time.Now()
	for i := 0; i < 10; i++ {
		g.Go(sleep)
	}

	require.Error(t, g.Wait())
	require.WithinDuration(t, now, time.Now(), 3*time.Second)
}

