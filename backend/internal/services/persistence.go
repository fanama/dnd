package services

import (
	"sync"

	"dnd-backend/internal/repository"
)

// Persister serializes every database write onto a single background
// goroutine so the game loop never blocks on SQLite writes.
//
// Writes are coalesced per key: when a write for a key is still pending and a
// newer snapshot for the same key arrives, the pending one is replaced (last
// write wins). A burst of actions on the same character — or on the world —
// therefore costs a single database write instead of one write per action.
// Enqueue never blocks the caller, so a slow disk can no longer stall the
// game-loop mutex.
type Persister struct {
	repo    *repository.SQLiteRepository
	mu      sync.Mutex
	pending map[string]func()
	wake    chan struct{}
	done    chan struct{}
	wg      sync.WaitGroup
	closed  bool
}

func NewPersister(repo *repository.SQLiteRepository) *Persister {
	p := &Persister{
		repo:    repo,
		pending: make(map[string]func()),
		wake:    make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
	p.wg.Add(1)
	go p.loop()
	return p
}

func (p *Persister) loop() {
	defer p.wg.Done()
	for {
		select {
		case <-p.done:
			p.drain()
			return
		case <-p.wake:
			p.drain()
		}
	}
}

// drain executes every pending job. Jobs were deduplicated at enqueue time,
// so each key runs at most once per drain round.
func (p *Persister) drain() {
	for {
		p.mu.Lock()
		if len(p.pending) == 0 {
			p.mu.Unlock()
			return
		}
		jobs := p.pending
		p.pending = make(map[string]func())
		p.mu.Unlock()
		for _, job := range jobs {
			job()
		}
	}
}

// Enqueue queues a write job under key. A pending job with the same key is
// replaced by the newer one, so the freshest snapshot always wins. It never
// blocks; jobs enqueued after Close are dropped.
func (p *Persister) Enqueue(key string, job func()) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.pending[key] = job
	p.mu.Unlock()
	select {
	case p.wake <- struct{}{}:
	default:
	}
}

// Close flushes pending writes and stops the writer goroutine. Idempotent.
func (p *Persister) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()
	close(p.done)
	p.wg.Wait()
}
