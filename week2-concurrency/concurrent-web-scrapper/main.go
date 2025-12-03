package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ScrapperResult struct {
	URL        string
	StatusCode int
	BodyLength int
	Duration   time.Duration
	Error      error
}

type Scrapper struct {
	numWorkers int
	urls       chan string
	results    chan ScrapperResult
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	client     *http.Client
	limiter    *rate.Limiter
}

func NewScrapper(numWorkers int, timeout time.Duration, rps int) *Scrapper {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scrapper{
		client: &http.Client{
			Timeout: timeout,
		},
		numWorkers: numWorkers,
		urls:       make(chan string, 1000),
		results:    make(chan ScrapperResult, 100),
		ctx:        ctx,
		cancel:     cancel,
		limiter:    rate.NewLimiter(rate.Limit(rps), rps),
	}
}

func (sc *Scrapper) Start() {
	for i := range sc.numWorkers {
		sc.wg.Add(1)
		go sc.Worker(i)
	}
}

func (sc *Scrapper) Worker(workerID int) {
	defer sc.wg.Done()

	for {
		select {
		case url, ok := <-sc.urls:
			if !ok {
				return
			}

			result := sc.ScrapeURL(url)

			select {
			case sc.results <- result:
			case <-sc.ctx.Done():
				return
			}
		case <-sc.ctx.Done():
			return
		}
	}
}

func (sc *Scrapper) ScrapeURL(url string) ScrapperResult {
	start := time.Now()

	result := ScrapperResult{
		URL: url,
	}

	sc.limiter.Wait(sc.ctx)

	resp, err := sc.client.Get(url)
	if err != nil {
		result.Duration = time.Since(start)
		result.Error = err
		return result
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Duration = time.Since(start)
		result.Error = err
		return result
	}

	result.StatusCode = resp.StatusCode
	result.BodyLength = len(body)
	result.Duration = time.Since(start)
	return result
}

func (sc *Scrapper) Submit(url string) error {
	select {
	case sc.urls <- url:
		return nil
	case <-sc.ctx.Done():
		return fmt.Errorf("Scrapper Shutting down")
	}
}

func (sc *Scrapper) Stop() {
	fmt.Println("==========Stopping Scrapper==========")
	close(sc.urls)
	sc.wg.Wait()
	close(sc.results)
	sc.cancel()
	fmt.Println("==========Scrapper Stopped===========")
}

func (sc *Scrapper) Results() <-chan ScrapperResult {
	return sc.results
}

func generateURLs(count int) []string {
	urls := make([]string, count)
	for i := range count {
		urls[i] = fmt.Sprintf("https://httpbin.org/delay/%d", i%5)
	}
	return urls
}

func main() {
	fmt.Println("Web Scrapper")

	scrapper := NewScrapper(50, 60*time.Second, 50)

	scrapper.Start()

	urls := generateURLs(1000)

	submitDone := make(chan struct{})
	go func() {
		for _, url := range urls {
			if err := scrapper.Submit(url); err != nil {
				fmt.Println("Error Submitting URL: ", url, ", Error: ", err)
			}
		}

		close(submitDone)
	}()

	go func() {
		<-submitDone
		scrapper.Stop()
	}()

	var (
		successCount  int
		failedCount   int
		totalDuration time.Duration
	)

	fmt.Println("Scrapping Results: ")
	fmt.Println(strings.Repeat("=", 80))

	for result := range scrapper.Results() {
		if result.Error != nil {
			failedCount++
			fmt.Printf("❌ %s - Error: %v\n", result.URL, result.Error)
		} else {
			successCount++
			totalDuration += result.Duration
			fmt.Printf("✅ %s - Status %d, Size: %d bytes, Time: %v\n", result.URL, result.StatusCode, result.BodyLength, result.Duration)
		}
	}

	// print statistics
	fmt.Println()
	fmt.Println("Statistics:")
	fmt.Printf("  Total URLs: %d \n", len(urls))
	fmt.Printf("  Success: %d\n", successCount)
	fmt.Printf("  Failed: %d\n", failedCount)

	if successCount > 0 {
		avgDuration := totalDuration / time.Duration(successCount)
		fmt.Printf("  Average Duration: %v\n", avgDuration)
	}

	fmt.Println("\nScrapping Complete")
}
