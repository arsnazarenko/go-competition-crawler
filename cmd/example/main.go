package main

import (
	"context"
	"fmt"
	"go-competiotion-crawler/internal/worker_pool"
	"io"
	"net/http"
	"time"
)

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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pool := worker_pool.NewWorkerPoolWithCapacity[string](8, 200)

	urls := [...]string{
		"www.zoom.com",
		"github.com",
		"gobyexample.com",
		"web.telegram.org",
		"gitflic.ru",
	}

	handles := make([]worker_pool.Handle[string], 0, 200)
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
