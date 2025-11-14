module github.com/blimu-dev/blimu-cli

go 1.25

replace github.com/blimu-dev/sdk-gen => ../sdk-gen

replace github.com/blimu-dev/blimu-go => ../blimu-go

replace github.com/blimu-dev/blimu-platform-go => ../blimu-platform-go

require (
	github.com/blimu-dev/blimu-go v0.0.0-00010101000000-000000000000
	github.com/blimu-dev/blimu-platform-go v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.9.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
