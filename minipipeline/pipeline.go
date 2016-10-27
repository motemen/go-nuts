package minipipeline

import (
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type Set struct {
	sync.Mutex
	Pipelines []*Pipeline
}

type Pipeline struct {
	Name  string
	Steps []*Step
	Err   error
}

type Step struct {
	Name       string
	StartedAt  time.Time
	FinishedAt time.Time
	Err        error

	progressDone uint32
	progressAll  uint32

	progressMu sync.Mutex
}

func (s *Step) ProgressDone(delta uint32) {
	s.progressMu.Lock()
	defer s.progressMu.Unlock()

	s.progressDone += delta
}

func (s *Step) ProgressAll(delta uint32) {
	s.progressMu.Lock()
	defer s.progressMu.Unlock()

	s.progressAll += delta
}

func (s *Set) Pipeline(name string) *Pipeline {
	s.Lock()
	defer s.Unlock()

	p := &Pipeline{Name: name}
	s.Pipelines = append(s.Pipelines, p)

	return p
}

func (p *Pipeline) Step(name string, fn func() error) error {
	if p.Err != nil {
		return p.Err
	}

	step := &Step{Name: name}
	p.Steps = append(p.Steps, step)

	log.Printf("%s {", name)

	step.StartedAt = time.Now()
	err := fn()
	step.FinishedAt = time.Now()

	log.Printf("} // %s", name)

	step.Err = err
	p.Err = err

	return err
}

func (p *Pipeline) Current() *Step {
	if len(p.Steps) == 0 {
		return nil
	}

	step := p.Steps[len(p.Steps)-1]
	if step.FinishedAt.IsZero() {
		// step is running
		return step
	}

	return nil
}

type ProgressGroup struct {
	step *Step
	*errgroup.Group
}

func (s *Step) ProgressGroup() ProgressGroup {
	return ProgressGroup{
		step:  s,
		Group: new(errgroup.Group),
	}
}

func (s *Step) String() string {
	if s.FinishedAt.IsZero() {
		// running
		s.progressMu.Lock()
		defer s.progressMu.Unlock()

		if s.progressDone != 0 || s.progressAll != 0 {
			return fmt.Sprintf("● %d/%d", s.progressDone, s.progressAll)
		} else {
			return "●"
		}
	} else if s.Err != nil {
		return fmt.Sprintf("✗ %s", s.Err)
	} else {
		return fmt.Sprintf("✓ %s", s.FinishedAt.Sub(s.StartedAt))
	}
}

func (pg *ProgressGroup) Go(fn func() error) {
	pg.step.ProgressAll(1)
	pg.Group.Go(func() error {
		defer pg.step.ProgressDone(1)
		return fn()
	})
}
