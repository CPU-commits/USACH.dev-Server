package utils

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/CPU-commits/USACH.dev-Server/res"
	"golang.org/x/sync/semaphore"
)

type key int

const (
	keyPrincipalID key = iota
)

type Context struct {
	Ctx    *context.Context
	Cancel context.CancelFunc
	Key    key
}

func Concurrency(
	wight int64,
	count int,
	do func(index int, ctx *Context),
) *res.ErrorRes {
	// Check if exists all users
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(wight)
	// Ctx with cancel if error
	ctx, cancel := context.WithCancel(context.Background())
	// Ctx error
	ctx = context.WithValue(ctx, keyPrincipalID, nil)

	wg.Add(count)
	for i := 0; i < count; i++ {
		if err := sem.Acquire(ctx, 1); err != nil {
			wg.Done()
			// Close go routines
			cancel()
			if errors.Is(err, context.Canceled) {
				if errRes := ctx.Value(keyPrincipalID); errRes != nil {
					return errRes.(*res.ErrorRes)
				}
			}
			return &res.ErrorRes{
				Err:        err,
				StatusCode: http.StatusBadRequest,
			}
		}
		// Wrapper
		go func(wg *sync.WaitGroup, index int) {
			defer wg.Done()
			do(index, &Context{
				Ctx:    &ctx,
				Cancel: cancel,
				Key:    keyPrincipalID,
			})
			// Free semaphore
			sem.Release(1)
		}(&wg, i)
	}
	// Close all
	wg.Wait()
	cancel()
	// Catch error
	if err := ctx.Value(keyPrincipalID); err != nil {
		return err.(*res.ErrorRes)
	}
	return nil
}
