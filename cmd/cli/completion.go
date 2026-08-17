package main

import (
	"flag"
	"fmt"
	"log"
)

const bashCompletion = `_outpipe_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts="login open create list inspect start stop revoke http tcp health completion version"
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}
complete -F _outpipe_completion outpipe
`

const zshCompletion = `#compdef outpipe
_outpipe() {
    local -a commands
    commands=(
        'login:Log in to Outpipe'
        'open:Open a tunnel'
        'create:Create a new tunnel'
        'list:List active tunnels'
        'inspect:Inspect tunnel details'
        'start:Start a tunnel'
        'stop:Stop a tunnel'
        'revoke:Revoke a tunnel'
        'http:Open an HTTP tunnel'
        'tcp:Open a TCP tunnel'
        'health:Check relay and API health'
        'completion:Generate shell completion'
        'version:Print version'
    )
    _describe -t commands 'command' commands
}
_outpipe "$@"
`

const fishCompletion = `complete -c outpipe -n "__fish_use_subcommand" -a "login open create list inspect start stop revoke http tcp health completion version"
`

func runCompletion(args []string) {
	flags := flag.NewFlagSet("completion", flag.ExitOnError)
	shell := flags.String("shell", "bash", "shell type: bash, zsh, fish")
	_ = flags.Parse(args)

	switch *shell {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	case "fish":
		fmt.Print(fishCompletion)
	default:
		log.Fatalf("unsupported shell %q; supported: bash, zsh, fish", *shell)
	}
}
