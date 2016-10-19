package minipipeline

import (
	"log"
	"sync"
	"time"
)

type Set struct {
	mu        sync.Mutex
	pipelines map[string]*Pipeline
}

type Pipeline struct {
	mu    sync.Mutex
	id    string
	steps []*Step
	err   error
}

type Step struct {
	Name       string
	StartedAt  time.Time
	FinishedAt time.Time
	Err        error
}

func (s *Set) Pipeline(id string) *Pipeline {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.pipelines == nil {
		s.pipelines = map[string]*Pipeline{}
	}

	if p, ok := s.pipelines[id]; ok {
		return p
	}

	p := &Pipeline{
		id:    id,
		steps: []*Step{},
	}
	s.pipelines[id] = p
	return p
}

/*
func (s *Set) WriteTo(w io.Writer) {
	table := tablewriter.NewWriter(w)

	headers := []string{}
	for _, p := range s.pipelines {
		for i := len(headers); i < len(p.steps); i++ {
			headers = append(headers, p.steps[i].Name)
		}
	}
	table.SetHeader(append([]string{"id"}, headers...))

	for id, p := range s.pipelines {
		row := make([]string, 1+len(p.steps))
		row[0] = id
		for i, s := range p.steps {
			if s.FinishedAt.IsZero() {
				row[i+1] = "●"
			} else if s.Err != nil {
				row[i+1] = "✗ " + s.Err.Error()
			} else {
				row[i+1] = "✓ " + s.FinishedAt.Sub(s.StartedAt).String()
			}
		}
		table.Append(row)
	}

	table.Render()
}
*/

func (p *Pipeline) Step(name string, fn func() error) {
	if p.err != nil {
		return
	}

	step := &Step{
		Name: name,
	}

	p.mu.Lock()
	p.steps = append(p.steps, step)
	p.mu.Unlock()

	log.Printf("[%s] %s {", p.id, name)

	step.StartedAt = time.Now()
	err := fn()
	step.FinishedAt = time.Now()

	log.Printf("[%s] } // %s err=%v", p.id, name, err)

	if err != nil {
		step.Err = err

		p.mu.Lock()
		if p.err == nil {
			p.err = err
		}
		p.mu.Unlock()
	}
}

func (p *Pipeline) Err() error {
	return p.err
}
