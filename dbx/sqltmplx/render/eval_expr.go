package render

import (
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func defaultExprRun(program *vm.Program, env any) (any, error) {
	return expr.Run(program, env)
}
