package subscription

type Action string

const (
	Subscribe   Action = "subscribe"
	Unsubscribe Action = "unsubscribe"
)

type Request struct {
	Action  Action   `json:"action"`
	Symbols []string `json:"symbols"`
}
