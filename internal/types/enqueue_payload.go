package types

type EnqueuePayload struct {
	ID      string `json:"id,omitempty"`
	Command string `json:"command"`
}
