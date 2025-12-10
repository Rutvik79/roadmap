package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	lastSeen time.Time
	count    int
}

type RateLimiter struct {
	visitors map[string]*visitor // IP → visitor data
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	// create a new Rate limiter
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}

	// cleanup old  visitors
	go rl.cleanupVisitors()

	// return ratelimiter
	return rl
}

// Cleanup func runs in the background as a go routine
// it deletes all the old visitors who have visited our server
// and the time since > window length
func (rl *RateLimiter) cleanupVisitors() {
	for {
		// Before every iteration it sleeps for a minute
		time.Sleep(time.Minute)

		// lock to avoid race condition when update the map
		rl.mu.Lock()

		// iterate over the visitors map
		for ip, v := range rl.visitors {
			// if time since last visit of the ip > window,
			// then delete client ip from visitors map
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		// release the lock
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the client ip
		ip := c.ClientIP()

		// using locks to update the map safely (without race conditions)when there are multiple concurrent request to the server
		rl.mu.Lock()
		// check if the client ip is present in visitors
		v, exists := rl.visitors[ip]

		// case 1: Firest request from this ip
		// if not present, that means client is first time visitor
		// put client into visitors map
		// release lock and call next middleware
		if !exists {
			rl.visitors[ip] = &visitor{
				lastSeen: time.Now(),
				count:    1,
			}

			rl.mu.Unlock()
			c.Next() //Allow
			return
		}

		// case 2: Window expired (reset counter)
		// if time since the ip was last seen > ratelimiters window
		// reset the count to and last seen to time.now
		// release lock and call next middleware
		if time.Since(v.lastSeen) > rl.window {
			v.count = 1
			v.lastSeen = time.Now()
			rl.mu.Unlock()
			c.Next() //Allow
			return
		}

		// case 3: Limit reached
		// if request count >= rate limiters limit then send rate limit exceeded message
		// release lock and return the response in the middleware itself dont process the request
		if v.count >= rl.limit {
			rl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate Limit exceeded",
			})
			c.Abort() //Block request
			return
		}

		// case 4: Within limit
		// increment the request count on every request
		// set the last seen of the client ip
		// release lock and call next middleware
		v.count++
		v.lastSeen = time.Now()
		rl.mu.Unlock()

		c.Next() // Allow
	}
}
