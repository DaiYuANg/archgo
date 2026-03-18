package render

import (
	"fmt"

	"github.com/expr-lang/expr/vm"
)

var exprRunner func(program *vm.Program, env any) (any, error)

func init() {
	exprRunner = defaultExprRun
}

func exprRun(program *vm.Program, env any) (any, error) {
	if exprRunner == nil {
		return nil, fmt.Errorf("sqltmplx: expr runner is nil")
	}
	return exprRunner(program, env)
}
