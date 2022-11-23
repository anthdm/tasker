package tasker

import (
	"context"
	"testing"
	"time"
)

func TestSpawnWithTimeout(t *testing.T) {
	fn := func(ctx context.Context) (struct{}, error) {
		time.Sleep(time.Millisecond * 2)
		return struct{}{}, nil
	}
	duration := time.Millisecond * 1
	task := SpawnWithTimeout(fn, duration)

	_, err := task.Await()
	if err != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded error but got %v", err)
	}
}

func TestSpawnTaskAndCancel(t *testing.T) {
	fn := func(ctx context.Context) (struct{}, error) {
		return struct{}{}, nil
	}
	task := Spawn(fn)
	task.Cancel()

	_, err := task.Await()
	if err != context.Canceled {
		t.Errorf("expected cancelation error but got %v", err)
	}
}

func TestSpawnTask(t *testing.T) {
	expected := 1
	fn := func(ctx context.Context) (int, error) {
		return expected, nil
	}

	task := Spawn(fn)
	val, err := task.Await()
	if err != nil {
		t.Error(err)
	}
	if val != expected {
		t.Errorf("expected %d but got %d", expected, val)
	}
}
