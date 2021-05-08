package main

type payload struct {
	Name       string            `json:"-"`
	State      string            `json:"state,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type payloads struct {
	data []payload
}

func (pp *payloads) Add(p payload) *payloads {
	pp.data = append(pp.data, p)
	return pp
}

func SinglePayload(p payload) *payloads {
	return &payloads{data: []payload{p}}
}

func MultiplePayloads() *payloads {
	return &payloads{}
}
