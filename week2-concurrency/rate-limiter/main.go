package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RateLimiter struct {
	rate   int
	ticker *time.Ticker
	tokens chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewRateLimiter(requestPerSecond int) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		rate:   requestPerSecond,
		ticker: time.NewTicker(time.Second / time.Duration(requestPerSecond)),
		tokens: make(chan struct{}, requestPerSecond),
		ctx:    ctx,
		cancel: cancel,
	}

	// pre-fill the bucket with tokens
	for i := 0; i < requestPerSecond; i++ {
		rl.tokens <- struct{}{}
	}

	rl.wg.Add(1)
	go rl.refillTokens()

	return rl
}

func (rl *RateLimiter) refillTokens() {
	defer rl.wg.Done()

	for {
		select {
		case <-rl.ticker.C:
			select {
			case rl.tokens <- struct{}{}:
			default:
				// bucket full, skip
			}
		case <-rl.ctx.Done():
			return
		}
	}
}

// allow blocks until token is available
func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	case <-rl.ctx.Done():
		return false
	}
}

func (rl *RateLimiter) TryAllow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) Stop() {
	fmt.Println("======= Stopping Rate Limiter =======")
	rl.ticker.Stop()
	rl.cancel()
	rl.wg.Wait()
	close(rl.tokens)
	fmt.Println("======= Rate Limiter Stopped =======")
}

// basic rate limiting
func example1() {
	fmt.Println("======= Example 1. Basic Rate Limiting =======")

	limiter := NewRateLimiter(5) // 5 requests per second
	defer limiter.Stop()

	start := time.Now()

	for i := 1; i < 10; i++ {
		limiter.Allow()
		elapsed := time.Since(start)
		fmt.Printf("Request %d at %v\n", i, elapsed.Round(time.Millisecond))
	}
}

// Example 2: Rate limiter with workers
func example2() {
	fmt.Println("\n====== Example 2. Rate Limiters With Workers ======")
	limiter := NewRateLimiter(10) // 10 requests per second

	defer limiter.Stop()

	var wg sync.WaitGroup
	request := 30

	for i := 1; i <= request; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// wait for token
			// blocks execution here until next tick
			if !limiter.Allow() {
				// if false means context cancelled
				return
			}

			fmt.Printf("Processing request %d at %v\n", id, time.Now().Format("15:04:05.000"))
		}(i)
	}

	wg.Wait()
}

// Example 3: Non-blocking rate limiter
func example3() {
	fmt.Println("====== Example 3. Non-Blocking Rate Limiter ======")
	limiter := NewRateLimiter(3) // 3 Requests Per Second
	defer limiter.Stop()

	accepted, rejected := 0, 0

	for i := 1; i <= 10; i++ {
		if limiter.TryAllow() {
			accepted++
			fmt.Printf("✅ Request %d accepted\n", i)
		} else {
			rejected++
			fmt.Printf("❌ Request %d rejected\n", i)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\nAccepted: %d, Rejected: %d\n", accepted, rejected)
}

// Advanced: Token Bucket with Burst
type TockenBucket struct {
	capacity   int
	tokens     int
	rate       time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTockenBucket(capacity int, refillRate time.Duration) *TockenBucket {
	return &TockenBucket{
		capacity:   capacity,
		tokens:     capacity,
		rate:       refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TockenBucket) Allow(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed / tb.rate)

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	// check if we have enough tokens
	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

func example4() {
	fmt.Println("====== Example 4. Token Bucket with burst ======")

	bucket := NewTockenBucket(10, 100*time.Millisecond) // 10 tokens, Refill every 100ms

	fmt.Println("Bursts of 10 request.")
	for i := 1; i <= 10; i++ {
		if bucket.Allow(1) {
			fmt.Printf("✅ Request %d\n", i)
		} else {
			fmt.Printf("❌ Request %d (rate limited)\n", i)
		}
	}

	// wait for refill
	time.Sleep(500 * time.Millisecond)
	for i := 11; i <= 15; i++ {
		if bucket.Allow(1) {
			fmt.Printf("✅ Request %d\n", i)
		} else {
			fmt.Printf("❌ Request %d (rate limited)\n", i)
		}
	}
}

func main() {
	fmt.Println("============ Rate Limiter ============")
	// example1()
	// time.Sleep(1 * time.Second)

	// example2()
	// time.Sleep(1 * time.Second)

	// example3()
	// time.Sleep(1 * time.Second)

	example4()
	time.Sleep(1 * time.Second)

	fmt.Println("End of Program")
}
