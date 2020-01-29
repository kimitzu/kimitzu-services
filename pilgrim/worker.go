package pilgrim

import (
    "context"
    "time"
)

type Worker struct {
    Identifier string
    Manager    *Manager
    Type       string
    Started    bool
    JobQueue   chan *Job
    stop       chan interface{}
}

// Start to loop though jobs in the jobqueue
func (w *Worker) Start() {
    log := log.Sub("w-" + w.Identifier)
    go func() {
        w.Started = true
        log.Infof("Worker start: %v", w.Identifier)

        for job := range w.JobQueue {
            log.Verbosef("Execute Job: '%v'", job.Target)
            w.ExecuteJob(job)
        }

    }()
}

// ExecuteJob handles the job runtime sandbox which includes job execution time limits
// this ensures that a stray channel lock won't eventually stall the worker pool
func (w *Worker) ExecuteJob(job *Job) {
    log := log.Sub("wEx-" + w.Identifier)
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*340)
    success := make(chan interface{})
    var err error
    var response interface{}

    go func() {
        response, err = job.Execute()
        if err != nil {
            log.Errorf("Worker '%v' failed job: %v", w.Identifier, err)
            cancel()
        } else {
            success <- 1
        }
    }()

    select {
    case <-success:
        log.Verbosef("Worker '%v' successfully finished job '%v'", w.Identifier, job.Target)
        job.Response <- response
    case <-ctx.Done():
        if err == nil {
            log.Errorf("Worker '%v' failed job '%v': %v", w.Identifier, job.Target, "Job timeout exceeded")
        }
        job.Response <- nil
    }

}

func (w *Worker) Stop() {
    log.Infof("Signaling to stop: %v", w.Identifier)
    w.stop <- 1
}
