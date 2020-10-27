package rabbitmq

import "errors"

var (
	ErrClosed = errors.New("pool is closed")
)

// Pool ...
type Pool interface {
	Get() (interface{}, error)

	Put(interface{}) error

	Close(interface{}) error

	Release()

	Len() int
}
