package verify

type (
	Config struct {
		Rules struct {
			AllowedLicenses           []string `yaml:"allowed_licenses"`
			WhitelistedPackageSources []string `yaml:"whitelisted_package_sources"`
		} `yaml:"rules"`
	}
)
