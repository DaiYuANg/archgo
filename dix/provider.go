package dix

func Provider0[T any](fn func() T) ProviderFunc       { return func(c *Container) { ProvideT(c, fn) } }
func Provider1[T, D1 any](fn func(D1) T) ProviderFunc { return func(c *Container) { Provide1T(c, fn) } }
func Provider2[T, D1, D2 any](fn func(D1, D2) T) ProviderFunc {
	return func(c *Container) { Provide2T(c, fn) }
}
func Provider3[T, D1, D2, D3 any](fn func(D1, D2, D3) T) ProviderFunc {
	return func(c *Container) { Provide3T(c, fn) }
}
func Provider4[T, D1, D2, D3, D4 any](fn func(D1, D2, D3, D4) T) ProviderFunc {
	return func(c *Container) { Provide4T(c, fn) }
}
func Provider5[T, D1, D2, D3, D4, D5 any](fn func(D1, D2, D3, D4, D5) T) ProviderFunc {
	return func(c *Container) { Provide5T(c, fn) }
}
func Provider6[T, D1, D2, D3, D4, D5, D6 any](fn func(D1, D2, D3, D4, D5, D6) T) ProviderFunc {
	return func(c *Container) { Provide6T(c, fn) }
}
