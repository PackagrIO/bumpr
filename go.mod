module github.com/packagrio/bumpr

go 1.13

require (
	github.com/Masterminds/semver v1.5.0
	github.com/analogj/go-util v0.0.0-20200905200945-3b93d31215ae
	github.com/golang/mock v1.4.4
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/packagrio/go-common v0.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.4
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.2.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

//replace github.com/packagrio/go-common v0.0.1 => ../go-common
