package gh

const (
	KeySig      = "X-Hub-Signature-256"
	ContentType = "Content-Type"

	EventType = "X-Github-Event"

	DefaultNotFound = "Value not present"
	MaxBodyLength   = 25 * 1024 * 1024 // https://docs.github.com/en/webhooks/webhook-events-and-payloads?actionType=in_progress#payload-cap
)
