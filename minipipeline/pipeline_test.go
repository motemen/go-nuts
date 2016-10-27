package minipipeline

import (
	"fmt"
	"log"
	"testing"
)

func TestSet_Pipeline(t *testing.T) {
	var set Set
	p := set.Pipeline("test1")

	logSteps := func() {
		ss := make([]string, len(p.Steps))
		for i, s := range p.Steps {
			ss[i] = s.String()
		}
		log.Printf("%v", ss)
	}

	p.Step("step1", func() error {
		logSteps()
		return nil
	})
	logSteps()

	p.Step("step2", func() error {
		logSteps()
		pg := p.Current().ProgressGroup()
		for i := 0; i < 3; i++ {
			pg.Go(func() error {
				logSteps()
				return nil
			})
		}
		return pg.Wait()
	})
	logSteps()

	p.Step("step3", func() error {
		logSteps()
		return fmt.Errorf("stop")
	})
	logSteps()

	p.Step("step4", func() error {
		t.Error("this step should not be executed")
		return nil
	})
	logSteps()
}
