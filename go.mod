module github.com/benchkram/bob

go 1.16

require (
	github.com/benchkram/errz v0.0.0-20211210135050-f997ca868855
	github.com/cespare/xxhash/v2 v2.1.2
	github.com/charmbracelet/bubbles v0.10.3
	github.com/charmbracelet/bubbletea v0.20.0
	github.com/charmbracelet/lipgloss v0.5.0
	github.com/cli/cli v1.14.0
	github.com/compose-spec/compose-go v1.1.0
	github.com/docker/cli v20.10.16+incompatible
	github.com/docker/compose/v2 v2.3.3
	github.com/docker/docker v20.10.16+incompatible
	github.com/fatih/structs v1.1.0
	github.com/go-git/go-git/v5 v5.4.2
	github.com/google/go-cmp v0.5.8
	github.com/hashicorp/go-version v1.4.0
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mholt/archiver/v3 v3.5.1
	github.com/mitchellh/go-wordwrap v1.0.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	github.com/sanity-io/litter v1.5.5
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.11.0
	github.com/stretchr/testify v1.7.1
	github.com/whilp/git-urls v1.0.0
	github.com/xlab/treeprint v1.1.0
	github.com/yargevad/filepathx v1.0.0
	golang.org/x/sync v0.0.0-20220513210516-0976fa681c29
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20220512140231-539c8e751b99
	mvdan.cc/sh v2.6.4+incompatible
)

require (
	github.com/AlecAivazis/survey/v2 v2.3.4 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20220517143526-88bb52951d5b // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/containerd/containerd v1.6.4 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/distribution/distribution/v3 v3.0.0-20220516112011-c202b9b0d7b7 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/fvbommel/sortorder v1.0.2 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.10.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.15.4 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/sys/mount v0.3.2 // indirect
	github.com/moby/sys/signal v0.7.0 // indirect
	github.com/muesli/ansi v0.0.0-20211031195517-c9f0611b6c70 // indirect
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/opencontainers/runc v1.1.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.14 // indirect
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/prometheus/common v0.34.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/theupdateframework/notary v0.7.0 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xanzy/ssh-agent v0.3.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.32.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.32.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.32.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20220518034528-6f7dac969898 // indirect
	golang.org/x/net v0.0.0-20220517181318-183a9ca12b87 // indirect
	golang.org/x/sys v0.0.0-20220517195934-5e4e11fc645e // indirect
	golang.org/x/term v0.0.0-20220411215600-e5f449aeb171 // indirect
	golang.org/x/time v0.0.0-20220411224347-583f2d630306 // indirect
	google.golang.org/genproto v0.0.0-20220505152158-f39f71e6c8f3 // indirect
	google.golang.org/grpc v1.46.2 // indirect
	k8s.io/client-go v0.24.0 // indirect
)

replace (
	github.com/docker/cli => github.com/docker/cli v20.10.3-0.20210702143511-f782d1355eff+incompatible
	github.com/docker/docker => github.com/docker/docker v20.10.3-0.20220121014307-40bb9831756f+incompatible
)
