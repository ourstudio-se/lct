package deps

import (
	"context"
	"net/http"
)

const (
	LicenseUnknown = "#Unkonwn"

	DefaultParallelizationLevel = uint64(50)
)

type (
	LicenseResolverOptionSet struct {
		HTTPClient *http.Client
		Ctx        context.Context
	}
	LicenseResolverOption func(set *LicenseResolverOptionSet)

	LicenseResolverFunc func(context.Context, *DependencyNode) ([]string, error)
)

func (fn LicenseResolverFunc) Resolve(ctx context.Context, node *DependencyNode) ([]string, error) {
	return fn(ctx, node)
}

func WithHTTPClient(httpClient *http.Client) LicenseResolverOption {
	return func(set *LicenseResolverOptionSet) {
		set.HTTPClient = httpClient
	}
}

func WithBaseContext(ctx context.Context) LicenseResolverOption {
	return func(set *LicenseResolverOptionSet) {
		set.Ctx = ctx
	}
}

func NoOpLicenseResolver() LicenseResolverFunc {
	return func(_ context.Context, _ *DependencyNode) ([]string, error) {
		return nil, nil
	}
}
