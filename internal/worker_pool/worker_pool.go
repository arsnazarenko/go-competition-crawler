package worker_pool

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type Workers int
type Capacity int

var defaultWorkers = Workers(runtime.NumCPU())
var defaultCapacity = Capacity(32)

type WorkerPool[T any] interface {
	Submit(func() (T, error)) (Handle[T], error)
	Done()
}

type result[T any] struct {
	e error
	v T
}

type Handle[T any] struct {
	resultChan <-chan result[T]
	state      result[T]
	invoked    bool
}

func (h *Handle[T]) wait() {
	if !h.invoked {
		h.state = <-h.resultChan
	}
}

func (h *Handle[T]) Get() (T, error) {
	h.wait()
	return h.state.v, h.state.e
}

type workerPoolImpl[T any] struct {
	wg         *sync.WaitGroup
	workers    Workers
	submitChan chan func() (T, error)
	resultChan chan result[T]
}

func (w *workerPoolImpl[T]) Submit(proc func() (T, error)) (Handle[T], error) {
	w.submitChan <- proc
	return Handle[T]{
		resultChan: w.resultChan,
		state:      result[T]{},
		invoked:    false,
	}, nil
}

func (w *workerPoolImpl[T]) Done() {
	close(w.submitChan)
}

func (w *workerPoolImpl[T]) runWorker() error {
	defer w.wg.Done()
	for task := range w.submitChan {
		v, err := task()
		w.resultChan <- result[T]{
			e: err,
			v: v,
		}
	}
	return nil
}

func NewWorkerPool[T any](workers Workers) WorkerPool[T] {
	return NewWorkerPoolWithCapacity[T](workers, defaultCapacity)
}

func NewWorkerPoolWithCapacity[T any](workers Workers, capacity Capacity) WorkerPool[T] {
	pool := &workerPoolImpl[T]{
		wg:         &sync.WaitGroup{},
		workers:    workers,
		submitChan: make(chan func() (T, error), capacity),
		resultChan: make(chan result[T], capacity),
	}
	pool.wg.Add(int(workers))
	for range workers {
		go pool.runWorker()
	}
	go func() {
		pool.wg.Wait()
		close(pool.resultChan)
	}()
	return pool
}

func makeRequestTask(ctx context.Context, url string) func() (string, error) {
	return func() (string, error) {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return "", err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}
}

func Example1() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pool := NewWorkerPoolWithCapacity[string](8, 200)

	urls := [...]string{
		"www.zoom.com",
		"github.com",
		"gobyexample.com",
		"web.telegram.org",
		"gitflic.ru",
	}

	handles := make([]Handle[string], 0, 200)
	for i := range 200 {
		h, err := pool.Submit(makeRequestTask(ctx, "https://"+urls[i%len(urls)]))
		if err != nil {
			panic(fmt.Sprintf("submit error: %v", err))
		}
		handles = append(handles, h)
	}
	pool.Done()

	for i, h := range handles {
		v, err := h.Get()
		if err != nil {
			fmt.Printf("%d: error = %v\n", i, err)
		} else {
			fmt.Printf("%d: len = %d\n", i, len(v))
		}
	}

}
