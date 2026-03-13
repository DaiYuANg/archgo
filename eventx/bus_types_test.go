package eventx

import "context"

type userCreated struct {
	ID int
}

func nilContext() context.Context {
	return nil
}

func (e userCreated) Name() string {
	return "user.created"
}

