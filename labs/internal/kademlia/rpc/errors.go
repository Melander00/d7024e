package rpc

import "fmt"

type ErrTimeout struct {
}

func (e *ErrTimeout) Error() string {
	return fmt.Sprintf("Message timed out")
}
