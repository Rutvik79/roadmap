package main
import (
    "fmt"
    "context"
    "time"
    "sync"
    )

func workerPool(ctx context.Context, numWorkers int, jobs <-chan int, ag *sync.WaitGroup) {
    var wg sync.WaitGroup
    defer ag.Done()
    
    for i := 1; i <= numWorkers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            for {
                select {
                case <-ctx.Done():
                    fmt.Printf("Worker %d stopped: %v\n", id, ctx.Err())
                    return
                case job, ok := <-jobs:
                    if !ok {
                        fmt.Printf("Worker %d: no more jobs\n", id)
                        return
                    }
                    fmt.Printf("Worker %d processing job %d\n", id, job)
                    time.Sleep(500 * time.Millisecond)
                }
            }
        }(i)
    }
    
    wg.Wait()
    fmt.Println("All workers stopped")
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    var wg sync.WaitGroup
    jobs := make(chan int, 10)
    
    // Start worker pool
    wg.Add(1)
    go workerPool(ctx, 3, jobs, &wg)
    
    // Send jobs
    for i := 1; i <= 20; i++ {
        jobs <- i
        time.Sleep(200 * time.Millisecond)
    }
    close(jobs)
    wg.Wait()
    
    // time.Sleep(5 * time.Second)
}