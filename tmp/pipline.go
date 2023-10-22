package main

import (
	"errors"
	"fmt"
	"time"
)

// Pipeline is a struct that contains a list of stages to execute in order
type Pipeline struct {
	Stages []*Stage
}

// Stage is a struct that contains a function to execute and an error channel to return any errors
type Stage struct {
	Fn           func(interface{}) (interface{}, error)
	RollbackFn   func(interface{}) error
	RetryCount   int
	RetryOnError bool
}

// Run runs each stage in the pipeline in order, passing the output of each stage to the input of the next stage
func (p *Pipeline) Run(input interface{}) (interface{}, error) {
	var err error
	var output interface{} = input

	for _, s := range p.Stages {
		// Retry loop
		for retry := 0; retry <= s.RetryCount; retry++ {

			// Execute function
			output, err = s.Fn(output)

			// Check for error
			if err != nil {
				// If RetryOnError is true, retry the stage
				if s.RetryOnError && retry < s.RetryCount {
					continue
				}

				// Rollback
				for i := range p.Stages[:len(p.Stages)-1] {
					rollbackErr := p.Stages[len(p.Stages)-1-i].RollbackFn(output)
					if rollbackErr != nil {
						return nil, rollbackErr
					}
				}

				return nil, err
			}

			// No error, move on to next stage
			break
		}
	}

	return output, nil
}

// NewStage creates a new stage with the given function and retry count
func NewStage(fn func(interface{}) (interface{}, error), rollbackFn func(interface{}) error, retryCount int, retryOnError bool) *Stage {
	return &Stage{
		Fn:           fn,
		RollbackFn:   rollbackFn,
		RetryCount:   retryCount,
		RetryOnError: retryOnError,
	}
}

// NewPipeline creates a new pipeline with the given stages
func NewPipeline(stages ...*Stage) *Pipeline {
	p := &Pipeline{
		Stages: stages,
	}
	return p
}

func main() {
	// Create stages
	stage1 := NewStage(func(input interface{}) (interface{}, error) {
		output := input.(int) + 1
		fmt.Println("Stage 1: ", output)
		return output, nil
	}, func(input interface{}) error {
		fmt.Println("Stage 1 Rollback: ", input)
		return nil
	}, 0, true)

	stage2 := NewStage(func(input interface{}) (interface{}, error) {
		output := input.(int) + 2
		fmt.Println("Stage 2: ", output)
		time.Sleep(time.Second) // Simulate a long running stage
		return output, nil
	}, func(input interface{}) error {
		fmt.Println("Stage 2 Rollback: ", input)
		return nil
	}, 2, true)

	stage3 := NewStage(func(input interface{}) (interface{}, error) {
		output := input.(int) + 3
		fmt.Println("Stage 3: ", output)
		return output, errors.New("error in stage 3")
	}, func(input interface{}) error {
		fmt.Println("Stage 3 Rollback: ", input)
		return nil
	}, 0, false)

	stage4 := NewStage(func(input interface{}) (interface{}, error) {
		output := input.(int) + 4
		fmt.Println("Stage 4: ", output)
		return output, nil
	}, func(input interface{}) error {
		fmt.Println("Stage 4 Rollback: ", input)
		return nil
	}, 0, true)

	// Create pipeline
	p := NewPipeline(stage1, stage2, stage3, stage4)

	// Run pipeline
	output, err := p.Run(0)
	if err != nil {
		fmt.Println("Pipeline failed: ", err)
	} else {
		fmt.Println("Pipeline succeeded: ", output)
	}
}
