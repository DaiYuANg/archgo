package dix

// Validate validates the immutable app spec and current module graph.
func (a *App) Validate() error {
	_, err := newBuildPlan(a)
	return err
}
