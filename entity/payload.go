package entity

type Payload struct {
	Icon       string                 `json:"icon,omitempty"`
	State      interface{}            `json:"state,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

func NewPayload() *Payload {
	return &Payload{
		Attributes: make(map[string]interface{}),
	}
}
