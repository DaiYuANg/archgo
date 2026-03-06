package authx

import (
	"fmt"
	"log/slog"

	collectionlist "github.com/DaiYuANg/arcgo/collectionx/list"
)

type policySourceChain struct {
	sources *collectionlist.ConcurrentList[PolicySource]
	logger  *slog.Logger
}

func newPolicySourceChain(sources ...PolicySource) (*policySourceChain, error) {
	chain := &policySourceChain{
		sources: collectionlist.NewConcurrentList[PolicySource](),
		logger:  normalizeLogger(nil).With("component", "authx.policy-source-chain"),
	}
	if len(sources) == 0 {
		return chain, nil
	}

	if err := chain.Set(sources...); err != nil {
		return nil, err
	}
	return chain, nil
}

func (s *policySourceChain) SetLogger(logger *slog.Logger) {
	if s == nil {
		return
	}
	s.logger = normalizeLogger(logger).With("component", "authx.policy-source-chain")
}

func (s *policySourceChain) Set(sources ...PolicySource) error {
	if s == nil || s.sources == nil {
		return fmt.Errorf("%w: policy source chain is nil", ErrInvalidPolicy)
	}

	validatedSources, err := validateSources(sources)
	if err != nil {
		return err
	}

	s.sources.Clear()
	s.sources.Add(validatedSources...)
	s.logger.Debug("policy source chain replaced", "sources", len(validatedSources))
	return nil
}

func (s *policySourceChain) Add(source PolicySource) error {
	if s == nil || s.sources == nil {
		return fmt.Errorf("%w: policy source chain is nil", ErrInvalidPolicy)
	}
	if source == nil {
		return fmt.Errorf("%w: policy source is nil", ErrInvalidPolicy)
	}

	s.sources.Add(source)
	s.logger.Debug("policy source added", "source_type", fmt.Sprintf("%T", source))
	return nil
}

func (s *policySourceChain) All() []PolicySource {
	if s == nil || s.sources == nil {
		return nil
	}
	return s.sources.Values()
}

func validateSources(sources []PolicySource) ([]PolicySource, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("%w: policy sources are required", ErrInvalidPolicy)
	}

	validatedSources := make([]PolicySource, 0, len(sources))
	for _, source := range sources {
		if source == nil {
			return nil, fmt.Errorf("%w: policy source is nil", ErrInvalidPolicy)
		}
		validatedSources = append(validatedSources, source)
	}
	return validatedSources, nil
}
