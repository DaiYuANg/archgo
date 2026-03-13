package clientx

import "github.com/samber/lo"

// Apply applies non-nil function options to target.
func Apply[T any, O ~func(*T)](target *T, opts ...O) {
	lo.ForEach(opts, func(opt O, _ int) {
		if opt != nil {
			opt(target)
		}
	})
}
