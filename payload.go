package main

type payload struct {
	State      string            `json:"state,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func NewPayload() *payload {
	return &payload{
		Attributes: make(map[string]string),
	}
}
