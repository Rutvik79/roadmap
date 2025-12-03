package main

import (
	"fmt"
	"sync"
	"time"
)

func Player(name string, ball chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		hit, ok := <-ball
		if !ok { // channel closed
			fmt.Printf("%s Lost...\n", name)
			return
		}
		fmt.Printf("%s hit the ball\n", name)

		if hit >= 10 {
			fmt.Printf("%s Wins!\n", name)
			close(ball)
			return
		}

		time.Sleep(200 * time.Millisecond)

		// hit the ball back
		ball <- hit + 1
	}
}

func main() {
	var wg sync.WaitGroup
	ball := make(chan int)

	wg.Add(1)
	go Player("Alice", ball, &wg)
	wg.Add(1)
	go Player("Bob", ball, &wg)

	// Start the game
	ball <- 1

	// Wait for the game to finish
	// time.Sleep(3 * time.Second)
	wg.Wait()
	fmt.Println("Game finished!")
}
