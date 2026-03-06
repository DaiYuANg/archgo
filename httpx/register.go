package httpx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/lo"
)

// Route registers related handlers.
func Route[I, O any](s *Server, method, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return registerTypedWithPrefix(s, "", method, path, handler, operationOptions...)
}

// Get registers related handlers.
func Get[I, O any](s *Server, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return Route(s, MethodGet, path, handler, operationOptions...)
}

// Post registers related handlers.
func Post[I, O any](s *Server, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return Route(s, MethodPost, path, handler, operationOptions...)
}

// Put registers related handlers.
func Put[I, O any](s *Server, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return Route(s, MethodPut, path, handler, operationOptions...)
}

// Patch registers related handlers.
func Patch[I, O any](s *Server, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return Route(s, MethodPatch, path, handler, operationOptions...)
}

// Delete registers related handlers.
func Delete[I, O any](s *Server, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return Route(s, MethodDelete, path, handler, operationOptions...)
}

// Head registers related handlers.
func Head[I, O any](s *Server, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return Route(s, MethodHead, path, handler, operationOptions...)
}

// Options registers related handlers.
func Options[I, O any](s *Server, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return Route(s, MethodOptions, path, handler, operationOptions...)
}

// GroupRoute registers related handlers.
func GroupRoute[I, O any](g *Group, method, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	if g == nil || g.server == nil {
		return fmt.Errorf("%w: route group is nil", ErrRouteNotRegistered)
	}
	return registerTypedWithPrefix(g.server, g.prefix, method, path, handler, operationOptions...)
}

// GroupGet registers related handlers.
func GroupGet[I, O any](g *Group, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return GroupRoute(g, MethodGet, path, handler, operationOptions...)
}

// GroupPost registers related handlers.
func GroupPost[I, O any](g *Group, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return GroupRoute(g, MethodPost, path, handler, operationOptions...)
}

// GroupPut registers related handlers.
func GroupPut[I, O any](g *Group, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return GroupRoute(g, MethodPut, path, handler, operationOptions...)
}

// GroupPatch registers related handlers.
func GroupPatch[I, O any](g *Group, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return GroupRoute(g, MethodPatch, path, handler, operationOptions...)
}

// GroupDelete registers related handlers.
func GroupDelete[I, O any](g *Group, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	return GroupRoute(g, MethodDelete, path, handler, operationOptions...)
}

func registerTypedWithPrefix[I, O any](s *Server, prefix, method, path string, handler TypedHandler[I, O], operationOptions ...OperationOption) error {
	if s == nil {
		return fmt.Errorf("%w: server is nil", ErrRouteNotRegistered)
	}
	if handler == nil {
		return fmt.Errorf("%w: handler cannot be nil", ErrRouteNotRegistered)
	}

	fullPath := joinRoutePath(joinRoutePath(s.basePath, prefix), path)
	upperMethod := strings.ToUpper(strings.TrimSpace(method))
	if upperMethod == "" {
		return fmt.Errorf("%w: method is required", ErrRouteNotRegistered)
	}

	op := huma.Operation{
		OperationID: defaultOperationID(upperMethod, fullPath),
		Method:      upperMethod,
		Path:        fullPath,
	}
	lo.ForEach(operationOptions, func(apply OperationOption, _ int) {
		if apply != nil {
			apply(&op)
		}
	})

	validatedHandler := withInputValidation(s, handler)

	api := s.adapter.HumaAPI()
	if api == nil {
		return fmt.Errorf("%w: adapter %q does not expose Huma API", ErrAdapterNotFound, s.adapter.Name())
	}
	huma.Register[I, O](api, op, validatedHandler)

	s.addRoute(RouteInfo{
		Method:      upperMethod,
		Path:        fullPath,
		HandlerName: handlerName(handler),
		Comment:     op.Summary,
		Tags:        op.Tags,
	})

	s.printRoutesIfEnabled()
	return nil
}

func withInputValidation[I, O any](s *Server, handler TypedHandler[I, O]) TypedHandler[I, O] {
	if handler == nil || s == nil {
		return handler
	}

	return func(ctx context.Context, input *I) (out *O, err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				out = nil
				err = huma.Error500InternalServerError(fmt.Sprintf("panic in handler: %v", recovered))
			}
		}()

		if err = s.validateInput(input); err != nil {
			message := validationErrorMessage(err)
			return nil, huma.Error400BadRequest(message, err)
		}

		out, err = handler(ctx, input)
		if err != nil {
			return nil, toHumaError(err)
		}
		return out, nil
	}
}

func toHumaError(err error) error {
	if err == nil {
		return nil
	}

	var httpxErr *Error
	if errors.As(err, &httpxErr) {
		if httpxErr.Err != nil {
			return huma.NewError(httpxErr.Code, httpxErr.Message, httpxErr.Err)
		}
		return huma.NewError(httpxErr.Code, httpxErr.Message)
	}

	return err
}

func handlerName(fn any) string {
	v := reflect.ValueOf(fn)
	if !v.IsValid() || v.Kind() != reflect.Func {
		return "unknown"
	}
	if runtimeFn := runtime.FuncForPC(v.Pointer()); runtimeFn != nil {
		parts := strings.Split(runtimeFn.Name(), "/")
		return parts[len(parts)-1]
	}
	return "unknown"
}

func defaultOperationID(method, path string) string {
	cleanPath := strings.Trim(path, "/")
	if cleanPath == "" {
		cleanPath = "root"
	}
	cleanPath = strings.NewReplacer("/", "-", "{", "", "}", "", ":", "").Replace(cleanPath)
	return strings.ToLower(method) + "-" + cleanPath
}
