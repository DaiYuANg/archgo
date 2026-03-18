package dialect

type SQLite struct{}

func (SQLite) BindVar(_ int) string { return "?" }
func (SQLite) Name() string         { return "sqlite" }
