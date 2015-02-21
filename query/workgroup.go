package query

import (
	"log"
	"sync"
)

type result struct {
	Feed  *Feed
	Error error
}

type task interface{}

type queryTask struct {
	Query      *Query
	ResultChan chan *result
}

type WorkGroup struct {
	taskChan chan task
}

func NewWorkGroup(numWorkers int) *WorkGroup {
	taskChan := make(chan task)

	for i := 0; i < numWorkers; i++ {
		go func(num int) {
			for {
				task, ok := <-taskChan
				if !ok {
					return
				}

				switch actualTask := task.(type) {
				case queryTask:
					log.Printf("[%v] Fetching query: %v", num, actualTask.Query.URL())

					feed, err := actualTask.Query.FetchPage()
					if err != nil {
						actualTask.ResultChan <- &result{
							Error: err,
						}
					} else {
						actualTask.ResultChan <- &result{
							Feed: feed,
						}
					}
				default:
					log.Printf("[%v] Cannot handle task: %#v", num, actualTask)
				}

			}
		}(i)
	}

	return &WorkGroup{taskChan}
}

func (g *WorkGroup) NewQuery(project string) *Query {
	return newQuery(project, g)
}

func (g *WorkGroup) addQueryTaskWithOutput(query *Query, resultChan chan *result) {
	go func() {
		g.taskChan <- &queryTask{
			Query:      query,
			ResultChan: resultChan,
		}
	}()
}

func (g *WorkGroup) addQueryTask(query *Query) chan *result {
	resultChan := make(chan *result)
	g.addQueryTaskWithOutput(query, resultChan)
	return resultChan
}

func (g *WorkGroup) addQueryTasks(queries []*Query) chan []*result {
	multiResultChan := make(chan []*result)

	go func() {
		wg := new(sync.WaitGroup)

		results := make([]*result, len(queries))
		for i, query := range queries {
			wg.Add(1)
			go func(i int, query *Query) {
				results[i] = <-g.addQueryTask(query)
				wg.Done()
			}(i, query)
		}

		wg.Wait()

		multiResultChan <- results
	}()

	return multiResultChan
}
