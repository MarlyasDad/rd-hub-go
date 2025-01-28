package strategies

type EventsBuffer struct {
	depth  int
	cursor int
	buffer []string
}

func (b EventsBuffer) AddEvent(event StrategyEvent) error {
	// check len
	return nil
}
