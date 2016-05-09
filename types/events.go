package types

// Priority of an event.
type Priority byte

const (
	// PriNormal is normal priority.
	PriNormal Priority = iota // Must be zero to work as default
	// PriLow is low priority.
	PriLow
)

// AlertType is the type of alert.
type AlertType byte

const (
	// AlertInfo is alert level "info".
	AlertInfo AlertType = iota // Must be zero to work as default
	// AlertWarning is alert level "warning".
	AlertWarning
	// AlertError is alert level "error".
	AlertError
	// AlertSuccess is alert level "success".
	AlertSuccess
)

// Event represents an event, described at http://docs.datadoghq.com/guides/dogstatsd/
type Event struct {
	// Title of the event.
	Title string
	// Text of the event. Supports line breaks.
	Text string
	// DateHappened of the event. Unix epoch timestamp. Default is now when not specified in incoming metric.
	DateHappened int64
	// Hostname of the event.
	Hostname string
	// AggregationKey of the event, to group it with some other events.
	AggregationKey string
	// SourceTypeName of the event.
	SourceTypeName string
	// Tags of the event.
	Tags Tags
	// Priority of the event.
	Priority Priority
	// AlertType of the event.
	AlertType AlertType
}

// Events represents a list of events.
type Events []Event

// Each iterates over each event.
func (e Events) Each(f func(Event)) {
	for _, event := range e {
		f(event)
	}
}

// Clone performs a copy of events.
func (e Events) Clone() Events {
	destination := make(Events, len(e))
	copy(destination, e)
	return destination
}
