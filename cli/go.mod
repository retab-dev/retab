module github.com/retab-dev/retab/cli

go 1.25.0

require (
	github.com/retab-dev/retab/clients/go v0.0.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.9
	golang.org/x/image v0.41.0
	golang.org/x/term v0.43.0
)

require (
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
)

replace github.com/retab-dev/retab/clients/go => ../clients/go
