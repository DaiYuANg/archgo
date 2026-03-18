package dix

type SetupFunc struct {
	run  func(*Container, Lifecycle) error
	meta SetupMetadata
}

func (s SetupFunc) apply(c *Container, lc Lifecycle) error {
	if s.run == nil {
		return nil
	}
	return s.run(c, lc)
}

func Setup(fn func(*Container, Lifecycle) error) SetupFunc {
	return NewSetupFunc(fn, SetupMetadata{
		Label: "Setup",
	})
}
