package main

import (
	"log"
	"sync/atomic"
)

func (s *Server) startWorkers(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		workerID := i + 1

		s.workers.Add(1)

		go func() {
			defer s.workers.Done()

			for job := range s.jobs {
				log.Printf("worker %d processing job %s", workerID, job.ID)

				s.processFn(job)

				atomic.AddUint64(&s.processed, 1)

				log.Printf("worker %d finished job %s", workerID, job.ID)
			}
		}()
	}
}
