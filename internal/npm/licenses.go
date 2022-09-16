package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ourstudio-se/lct/internal/deps"
)

func NpmJsLicenseResolver(opts ...deps.LicenseResolverOption) deps.LicenseResolverFunc {
	set := deps.LicenseResolverOptionSet{
		HTTPClient: &http.Client{
			Timeout: time.Second * 2,
		},
	}

	for _, opt := range opts {
		opt(&set)
	}

	return func(_ context.Context, node *deps.DependencyNode) ([]string, error) {
		if node.IsRootNode() {
			return nil, nil
		}

		if len(node.Licenses) > 0 {
			return node.Licenses, nil
		}

		return resolveDependencyLicenses(set.Ctx, node.PackageName, node.PackageVersion, set)
	}
}

func resolveDependencyLicenses(ctx context.Context, pkg string, version string, set deps.LicenseResolverOptionSet) ([]string, error) {
	u := fmt.Sprintf("https://registry.npmjs.org/%s", pkg)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := set.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseLicensesNode(data, version)
}

func parseLicensesNode(b []byte, version string) ([]string, error) {
	var response struct {
		Versions map[string]struct {
			License string `json:"license"`
		} `json:"versions"`
	}

	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}

	v, ok := response.Versions[version]
	if !ok {
		return []string{deps.LicenseUnknown}, nil
	}

	if v.License == "" {
		return []string{deps.LicenseUnknown}, nil
	}

	return []string{v.License}, nil
}
