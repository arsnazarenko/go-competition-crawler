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
	WaitAllDone() []Handle[T]
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
	w.wg.Wait()
	close(w.resultChan)
}

// Blocking method. If you can get wait for all tasks done and get result, you cat simply call this method
// If you want to get results in real time, you can call this method in separate goroutine
func (w *workerPoolImpl[T]) WaitAllDone() []Handle[T] {
	w.Done()
	handles := []Handle[T]{}
	for h := range w.resultChan {
		handles = append(handles, Handle[T]{
			resultChan: w.resultChan,
			state:      h,
			invoked:    true,
		})
	}
	return handles
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
	wp := &workerPoolImpl[T]{
		wg:         &sync.WaitGroup{},
		workers:    workers,
		submitChan: make(chan func() (T, error), capacity),
		resultChan: make(chan result[T], capacity),
	}
	wp.wg.Add(int(workers))
	for range workers {
		go wp.runWorker()
	}
	return wp
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wp := NewWorkerPool[string](1)
	
	handles := []Handle[string]{}

	h, err := wp.Submit(makeRequestTask(ctx, "https://google.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://yandex.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://github.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://google.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://yandex.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://github.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://google.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://yandex.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://github.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://google.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://yandex.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://github.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://google.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://yandex.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://github.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://google.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://yandex.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://github.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://google.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://yandex.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://github.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://google.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://yandex.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://github.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://google.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://yandex.com"))
	handles = append(handles, h)
	h, err = wp.Submit(makeRequestTask(ctx, "https://github.com"))
	handles = append(handles, h)
	go wp.Done() // done submitting

	if err != nil {
		panic(err)
	}

	for i, h := range handles {
		v, err := h.Get()
		if err != nil {
			fmt.Printf("%d: error: %v\n", i, err)
		} else {
			fmt.Printf("len: %d\n", len(v))
		}
	}

}

func Example2() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wp := NewWorkerPoolWithCapacity[string](2, 64)

	wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	wp.Submit(makeRequestTask(ctx, "https://google.com"))
	wp.Submit(makeRequestTask(ctx, "https://yandex.ru"))
	wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	wp.Submit(makeRequestTask(ctx, "https://google.com"))
	wp.Submit(makeRequestTask(ctx, "https://yandex.ru"))
	wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	wp.Submit(makeRequestTask(ctx, "https://yandex.ru"))
	wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	wp.Submit(makeRequestTask(ctx, "https://google.com"))
	wp.Submit(makeRequestTask(ctx, "https://yandex.ru"))
	wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	wp.Submit(makeRequestTask(ctx, "https://google.com"))
	wp.Submit(makeRequestTask(ctx, "https://yandex.ru"))
	wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	wp.Submit(makeRequestTask(ctx, "https://google.com"))
	wp.Submit(makeRequestTask(ctx, "https://yandex.ru"))
	wp.Submit(makeRequestTask(ctx, "https://bing.com"))
	wp.Submit(makeRequestTask(ctx, "https://bing.com"))

	hs := wp.WaitAllDone()
	for i, h := range hs {
		v, err := h.Get()
		if err != nil {
			fmt.Printf("%d: error: %v\n", i, err)
		} else {
			fmt.Printf("%d: len: %d\n", i, len(v))
		}
	}
}
