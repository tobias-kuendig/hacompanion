package main

type payload struct {
	State      interface{}            `json:"state,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

func NewPayload() *payload {
	return &payload{
		Attributes: make(map[string]interface{}),
	}
}
