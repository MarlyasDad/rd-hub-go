package alor

import "errors"

var (
	ErrNewBarFound        = errors.New("new bar was found")
	ErrSubscriberNotFound = errors.New("subscriber not found")
	ErrNoAvailableHandler = errors.New("no available handler")
)
