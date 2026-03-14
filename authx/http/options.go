package authhttp

import "github.com/samber/lo"

// ApplyOptions applies non-nil option funcs to target.
func ApplyOptions[T any, O ~func(*T)](target *T, opts ...O) {
	lo.ForEach(opts, func(opt O, _ int) {
		if opt != nil {
			opt(target)
		}
	})
}
