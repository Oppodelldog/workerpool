package main

import (
	"fmt"
	"github.com/Oppodelldog/workerpool"
	"time"
)

type (
	dummyWorker struct{ isWorking bool }
)

func createDummyWorker() workerpool.Worker {
	return &dummyWorker{}
}

func (w *dummyWorker) Work(control workerpool.WorkControl) {
	w.isWorking = true
	go func() {
		fmt.Println(control.WorkLoad.(int))
		time.Sleep(300 * time.Millisecond)
		control.WorkResultChannel <- "done"
		w.isWorking = false
	}()
}

func (w *dummyWorker) IsWorking() bool {
	return w.isWorking
}

func main() {
	pool := workerpool.NewFireAndForget(createDummyWorker)

	for i := 0; i < 10000000000; i++ {
		time.Sleep(100 * time.Millisecond)
		pool.Work(i)
	}
}
