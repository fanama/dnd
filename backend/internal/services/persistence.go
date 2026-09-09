package services

import (
	"sync"

	"dnd-backend/internal/repository"
)

// Persister serializes every database write onto a single background
// goroutine so the game loop never blocks on SQLite writes.
type Persister struct {
	repo *repository.SQLiteRepository
	jobs chan func()
	wg   sync.WaitGroup
}

func NewPersister(repo *repository.SQLiteRepository) *Persister {
	p := &Persister{
		repo: repo,
		jobs: make(chan func(), 256),
	}
	p.wg.Add(1)
	go p.loop()
	return p
}

func (p *Persister) loop() {
	defer p.wg.Done()
	for job := range p.jobs {
		job()
	}
}

// Enqueue snapshots a write job for ordered execution by the writer goroutine.
func (p *Persister) Enqueue(job func()) {
	p.jobs <- job
}

// Close flushes pending writes and stops the writer goroutine.
func (p *Persister) Close() {
	close(p.jobs)
	p.wg.Wait()
}
