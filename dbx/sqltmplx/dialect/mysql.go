package dialect

type MySQL struct{}

func (MySQL) BindVar(_ int) string { return "?" }
