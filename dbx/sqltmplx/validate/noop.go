package validate

type Noop struct{}

func (Noop) Validate(string) error { return nil }
