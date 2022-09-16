package gomod

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ourstudio-se/lct/internal/deps"
	"golang.org/x/net/html"
)

func GoDevLicenseResolver(opts ...deps.LicenseResolverOption) deps.LicenseResolverFunc {
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

		return resolveDependencyLicenses(set.Ctx, node.DisplayName(), set)
	}
}

func resolveDependencyLicenses(ctx context.Context, pkg string, set deps.LicenseResolverOptionSet) ([]string, error) {
	u := fmt.Sprintf("https://pkg.go.dev/%s", pkg)
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

	return parseLicensesNode(data), nil
}

func parseLicensesNode(b []byte) []string {
	tkn := html.NewTokenizer(bytes.NewBuffer(b))
	isLicenseNode := false
	isLicenseHref := false

	for {
		node := tkn.Next()

		switch {
		case node == html.ErrorToken:
			return nil
		case node == html.StartTagToken:
			if isLicenseNode && tkn.Token().Data == "a" {
				isLicenseHref = true
				continue
			}

			for _, attr := range tkn.Token().Attr {
				if attr.Key == htmlLicenseAttrID && attr.Val == htmlLicenseAttrValue {
					isLicenseNode = true
					continue
				}
			}
		case node == html.TextToken:
			if isLicenseHref {
				data := strings.TrimSpace(tkn.Token().Data)
				items := strings.Split(data, ",")

				var licenses []string
				for _, item := range items {
					data := strings.TrimSpace(item)
					if data == "" {
						licenses = append(licenses, deps.LicenseUnknown)
						continue
					}

					licenses = append(licenses, data)
				}

				if len(licenses) == 0 {
					licenses = append(licenses, deps.LicenseUnknown)
				}

				return licenses
			}
		}
	}
}
