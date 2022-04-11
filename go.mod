module github.com/1xyz/pryrite

go 1.16

require (
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/briandowns/spinner v1.16.0
	github.com/c-bata/go-prompt v0.2.5
	github.com/charmbracelet/glamour v0.3.0
	github.com/creack/pty v1.1.13
	github.com/cristalhq/jwt/v3 v3.0.14
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/go-resty/resty/v2 v2.5.0
	github.com/google/uuid v1.2.0
	github.com/itchyny/timefmt-go v0.1.3
	github.com/jedib0t/go-pretty/v6 v6.1.0
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/mattn/go-shellwords v1.0.12
	github.com/mitchellh/go-ps v1.0.0
	github.com/muesli/reflow v0.2.1-0.20210115123740-9e1d0d53df68
	github.com/muesli/termenv v0.8.1
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.21.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/yuin/goldmark v1.3.3
	github.com/zalando/go-keyring v0.1.1
	go.etcd.io/bbolt v1.3.6
	go.uber.org/atomic v1.8.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/manifoldco/promptui => github.com/aardlabs/promptui v0.8.1-0.20210630201735-534be89033d6
	github.com/sanbornm/go-selfupdate v0.0.0-20210106163404-c9b625feac49 => github.com/aardlabs/go-selfupdate v0.0.0-20210615201232-2426e0201381
)
