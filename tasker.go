package tasker

import (
	"context"
	"time"
)

// TaskFunc is a func that can be scheduled in a goroutine.
// It should return a generic type T and an error.
type TaskFunc[T any] func(context.Context) (T, error)

type response[T any] struct {
	value T
	err   error
}

// Task will hold the response of the underlying TaskFunc.
// It will keep track whenever either the response of the TaskFunc
// is fullfilled or its context is canceled or timed out.
type Task[T any] struct {
	respch       chan response[T]
	ctx          context.Context
	cancel       context.CancelFunc
	parentCancel context.CancelFunc
}

// Await will block until either the TaskFunc returned a response or
// its context is canceled or timed out.
func (t *Task[T]) Await() (T, error) {
	select {
	case <-t.ctx.Done():
		var val T
		return val, t.ctx.Err()
	case resp := <-t.respch:
		return resp.value, resp.err
	}
}

// Cancel will cancel all underlying context routines. Use Cancel
// for canceling a TaskFunc or to prevent context leakage.
func (t *Task[T]) Cancel() {
	t.cancel()
	t.parentCancel()
}

// SpawnWithTimeout will execute a TaskFunc in a goroutine that will be automatically
// canceled after the given time.Duration.
func SpawnWithTimeout[T any](t TaskFunc[T], d time.Duration) *Task[T] {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	return spawn(ctx, cancel, t)
}

// Spawn executes a TaskFunc in another goroutine.
func Spawn[T any](t TaskFunc[T]) *Task[T] {
	ctx := context.Background()
	return spawn(ctx, func() {}, t)
}

func spawn[T any](ctx context.Context, parentCancel context.CancelFunc, t TaskFunc[T]) *Task[T] {
	respch := make(chan response[T])
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		val, err := t(ctx)
		respch <- response[T]{
			value: val,
			err:   err,
		}
		close(respch)
	}()

	return &Task[T]{
		respch:       respch,
		ctx:          ctx,
		cancel:       cancel,
		parentCancel: parentCancel,
	}
}
