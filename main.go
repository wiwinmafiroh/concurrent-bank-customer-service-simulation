package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Customer struct {
	id int
}

type Teller struct {
	id int
}

type Bank chan Teller
type Queue chan Customer

// ServeCustomer handles serving customers by tellers.
func (b Bank) ServeCustomer(queue Queue, spv *sync.WaitGroup) {
	defer spv.Done()

	for customer := range queue {
		tellerID := <-b
		fmt.Printf("Customer #%d is served by Teller #%d\n", customer.id, tellerID)

		// Simulate service time with a random duration between 0 and 15 seconds.
		serviceTime := time.Second * time.Duration(rand.Intn(16))
		time.Sleep(serviceTime)

		fmt.Printf("Customer #%d is done by Teller #%d\n", customer.id, tellerID)

		b <- tellerID
	}
}

func main() {
	// Initialize the random seed.
	rand.Seed(time.Now().UnixNano())

	// Create a "bank" channel with a capacity of 10 tellers.
	bank := make(Bank, 10)

	// Add tellers to the "bank" channel.
	for tellerID := 0; tellerID < cap(bank); tellerID++ {
		bank <- Teller{id: tellerID}
	}

	// Initialize WaitGroup to supervise goroutines.
	spv := &sync.WaitGroup{}

	// Create a "queue" channel for the customer queue.
	queue := make(Queue)

	// Start goroutines for each teller.
	for i := 0; i < cap(bank); i++ {
		spv.Add(1)
		go bank.ServeCustomer(queue, spv)
	}

	// Register interrupt signals (Ctrl+C or SIGTERM).
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Process the customer queue.
	for customerID := 0; ; customerID++ {
		select {
		case <-c:
			fmt.Println("The bank is preparing to close. Stopping the queue...")
			close(queue)
			spv.Wait()
			fmt.Println("Thank you, the bank is now closed.")
			return
		default:
		}

		fmt.Printf("Customer #%d is waiting...\n", customerID)

		// Send customers to the queue.
		queue <- Customer{id: customerID}
	}
}
