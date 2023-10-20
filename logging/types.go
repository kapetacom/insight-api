package logging

type LogEntry struct {
	Entity    string `json:"entity"`
	Timestamp int64  `json:"timestamp"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}
