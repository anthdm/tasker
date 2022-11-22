package main

import (
	"context"
	"fmt"
	"time"
)

type TaskFunc[T any] func(context.Context) (T, error)

type response[T any] struct {
	value T
	err   error
}

type Task[T any] struct {
	respch       chan response[T]
	ctx          context.Context
	cancel       context.CancelFunc
	parentCancel context.CancelFunc
}

func (t *Task[T]) Await() (T, error) {
	select {
	case <-t.ctx.Done():
		var val T
		return val, t.ctx.Err()
	case resp := <-t.respch:
		return resp.value, resp.err
	}
}

func (t *Task[T]) Cancel() {
	t.cancel()
	t.parentCancel()
}

type TaskOpts struct {
	Timeout time.Duration
}

func SpawnWithTimeout[T any](t TaskFunc[T], d time.Duration) *Task[T] {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	return spawn(ctx, cancel, t)
}

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

func main() {
	start := time.Now()
	userID := 69
	taskUserLikes := SpawnWithTimeout(fetchUserLikes(userID), time.Millisecond*600)

	likes, err := taskUserLikes.Await()
	if err != nil {
		fmt.Println("did not fetched user likes cause:", err)
	}

	fmt.Println("likes result: ", likes)
	fmt.Println("took: ", time.Since(start))
}

type User struct {
	Name string
}

func fetchUserLikes(userID int) TaskFunc[int] {
	return func(ctx context.Context) (int, error) {
		time.Sleep(time.Millisecond * 400)

		return 666, nil
	}
}
