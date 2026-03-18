package dix

type ProviderFunc struct {
	register func(*Container)
	meta     ProviderMetadata
}

func (p ProviderFunc) apply(c *Container) {
	if p.register != nil {
		p.register(c)
	}
}

func RawProvider(fn func(*Container)) ProviderFunc {
	return NewProviderFunc(fn, ProviderMetadata{
		Label: "RawProvider",
		Raw:   true,
	})
}

func Provider0[T any](fn func() T) ProviderFunc {
	return NewProviderFunc(
		func(c *Container) { ProvideT(c, fn) },
		ProviderMetadata{
			Label:  "Provider0",
			Output: TypedService[T](),
		},
	)
}

func Provider1[T, D1 any](fn func(D1) T) ProviderFunc {
	return NewProviderFunc(
		func(c *Container) { Provide1T(c, fn) },
		ProviderMetadata{
			Label:        "Provider1",
			Output:       TypedService[T](),
			Dependencies: []ServiceRef{TypedService[D1]()},
		},
	)
}

func Provider2[T, D1, D2 any](fn func(D1, D2) T) ProviderFunc {
	return NewProviderFunc(
		func(c *Container) { Provide2T(c, fn) },
		ProviderMetadata{
			Label:        "Provider2",
			Output:       TypedService[T](),
			Dependencies: []ServiceRef{TypedService[D1](), TypedService[D2]()},
		},
	)
}

func Provider3[T, D1, D2, D3 any](fn func(D1, D2, D3) T) ProviderFunc {
	return NewProviderFunc(
		func(c *Container) { Provide3T(c, fn) },
		ProviderMetadata{
			Label:        "Provider3",
			Output:       TypedService[T](),
			Dependencies: []ServiceRef{TypedService[D1](), TypedService[D2](), TypedService[D3]()},
		},
	)
}

func Provider4[T, D1, D2, D3, D4 any](fn func(D1, D2, D3, D4) T) ProviderFunc {
	return NewProviderFunc(
		func(c *Container) { Provide4T(c, fn) },
		ProviderMetadata{
			Label:        "Provider4",
			Output:       TypedService[T](),
			Dependencies: []ServiceRef{TypedService[D1](), TypedService[D2](), TypedService[D3](), TypedService[D4]()},
		},
	)
}

func Provider5[T, D1, D2, D3, D4, D5 any](fn func(D1, D2, D3, D4, D5) T) ProviderFunc {
	return NewProviderFunc(
		func(c *Container) { Provide5T(c, fn) },
		ProviderMetadata{
			Label:        "Provider5",
			Output:       TypedService[T](),
			Dependencies: []ServiceRef{TypedService[D1](), TypedService[D2](), TypedService[D3](), TypedService[D4](), TypedService[D5]()},
		},
	)
}

func Provider6[T, D1, D2, D3, D4, D5, D6 any](fn func(D1, D2, D3, D4, D5, D6) T) ProviderFunc {
	return NewProviderFunc(
		func(c *Container) { Provide6T(c, fn) },
		ProviderMetadata{
			Label:  "Provider6",
			Output: TypedService[T](),
			Dependencies: []ServiceRef{
				TypedService[D1](),
				TypedService[D2](),
				TypedService[D3](),
				TypedService[D4](),
				TypedService[D5](),
				TypedService[D6](),
			},
		},
	)
}
