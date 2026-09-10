package main

import "log"

func (s *Server) startWorkers(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		workerID := i + 1

		go func() {
			for job := range s.jobs {
				log.Printf("worker %d processing job %s", workerID, job.ID)

				s.processFn(job)

				log.Printf("worker %d finished job %s", workerID, job.ID)
			}
		}()
	}
}
