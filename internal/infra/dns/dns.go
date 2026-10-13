package dns

import "context"

type ChallengeSolver interface {
	Present(context.Context, string, string) error
	Cleanup(context.Context, string) error
	Name() string
}
