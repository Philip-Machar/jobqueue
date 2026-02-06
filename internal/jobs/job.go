package jobs

type Job struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Payload  interface{} `json:"payload"`
	Attempts int         `json:"attempts"`
}
