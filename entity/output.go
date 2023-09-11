package entity

import (
	"sync"
)

// Output{} is the resulting value of a Sensor that was run.
type Output struct {
	Payload *Payload
	Sensor  Sensor
}

type Outputs struct {
	Data []*Output
	lock sync.Mutex
}

func NewOutputs() Outputs {
	return Outputs{
		lock: sync.Mutex{},
		Data: make([]*Output, 0),
	}
}

func (o *Outputs) Add(output Output) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.Data = append(o.Data, &output)
}
