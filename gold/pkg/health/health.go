package health

import (
	"context"
	"net/http"
)

type Checker interface {
	Check(ctx context.Context) error
}

type FuncChecker func(ctx context.Context) error

func (f FuncChecker) Check(ctx context.Context) error {
	return f(ctx)
}

type namedChecker struct {
	Name    string
	Checker Checker
}

type Service struct {
	checkers []namedChecker
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Register(name string, checker Checker) {
	s.checkers = append(s.checkers, namedChecker{Name: name, Checker: checker})
}

func (s *Service) Check(ctx context.Context) (int, map[string]string) {
	components := make(map[string]string)
	allOK := true

	for _, c := range s.checkers {
		if err := c.Checker.Check(ctx); err != nil {
			components[c.Name] = err.Error()
			allOK = false
		} else {
			components[c.Name] = "ok"
		}
	}

	if !allOK {
		return http.StatusServiceUnavailable, components
	}
	return http.StatusOK, components
}
