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
		go func() {
			for {
				task, ok := <-taskChan
				if !ok {
					return
				}

				log.Printf("Fetching: %v", task.Query.URL())

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
		}()
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

func (g *WorkGroup) AddTasksUnordered(queries []*Query) chan []*Result {
	multiResultChan := make(chan []*Result)

	wg := new(sync.WaitGroup)
	resultChan := make(chan *Result)

	go func() {
		for _, query := range queries {
			wg.Add(1)
			g.AddTaskWithOutput(query, resultChan)
		}
	}()
	go func() {
		results := make([]*Result, 0, len(queries))
		for len(results) < len(queries) {
			result := <-resultChan
			results = append(results, result)
		}
		multiResultChan <- results
	}()

	return multiResultChan
}
