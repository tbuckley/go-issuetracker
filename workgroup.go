package main

import (
	"log"
	"sync"
)

type Result struct {
	Feed  *Feed
	Error error
}

type Task struct {
	Query      *Query
	ResultChan chan *Result
}

type WorkGroup struct {
	taskChan chan *Task
}

func NewWorkGroup(numWorkers int) *WorkGroup {
	taskChan := make(chan *Task)

	for i := 0; i < numWorkers; i++ {
		go func(num int) {
			for {
				task, ok := <-taskChan
				if !ok {
					return
				}

				log.Printf("[%v] Fetching: %v", num, task.Query.URL())

				feed, err := task.Query.FetchPage()
				if err != nil {
					task.ResultChan <- &Result{
						Error: err,
					}
				} else {
					task.ResultChan <- &Result{
						Feed: feed,
					}
				}
			}
		}(i)
	}

	return &WorkGroup{taskChan}
}

func (g *WorkGroup) AddTaskWithOutput(query *Query, resultChan chan *Result) {
	go func() {
		g.taskChan <- &Task{
			Query:      query,
			ResultChan: resultChan,
		}
	}()
}

func (g *WorkGroup) AddTask(query *Query) chan *Result {
	resultChan := make(chan *Result)
	g.AddTaskWithOutput(query, resultChan)
	return resultChan
}

func (g *WorkGroup) AddTasks(queries []*Query) chan []*Result {
	multiResultChan := make(chan []*Result)

	go func() {
		wg := new(sync.WaitGroup)

		results := make([]*Result, len(queries))
		for i, query := range queries {
			wg.Add(1)
			go func(i int, query *Query) {
				results[i] = <-g.AddTask(query)
				wg.Done()
			}(i, query)
		}

		wg.Wait()

		multiResultChan <- results
	}()

	return multiResultChan
}
