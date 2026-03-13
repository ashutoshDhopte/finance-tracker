package scheduler

import (
	"context"
	"log"
	"sync"
)

type Job struct {
	Name string
	Fn   func(ctx context.Context)
}

type Scheduler struct {
	jobs []Job
}

func New() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Add(name string, fn func(ctx context.Context)) {
	s.jobs = append(s.jobs, Job{Name: name, Fn: fn})
}

func (s *Scheduler) Start(ctx context.Context) {
	var wg sync.WaitGroup
	for _, job := range s.jobs {
		wg.Add(1)
		go func(j Job) {
			defer wg.Done()
			log.Printf("starting job: %s", j.Name)
			j.Fn(ctx)
		}(job)
	}
	wg.Wait()
}
