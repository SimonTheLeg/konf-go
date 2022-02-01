# Konf - Lightweight kubeconfig Manager

- [Konf - Lightweight kubeconfig Manager](#konf---lightweight-kubeconfig-manager)
  - [Why konf?](#why-konf)
  - [Installation](#installation)
  - [Usage](#usage)
  - [How does it work?](#how-does-it-work)
    - [kubeconfig management across shells](#kubeconfig-management-across-shells)
    - [zsh/bash-func-magic](#zshbash-func-magic)
  - [Contributing](#contributing)
    - [Usage of stdout and stderr](#usage-of-stdout-and-stderr)
    - [Tests](#tests)
  - [Ideas for Future Improvements](#ideas-for-future-improvements)

## Why konf?

- konf allows you to quickly switch between different kubeconfig files
- konf allows you to simultaneously use different kubeconfigs in different shells
- konf executes directly in your current shell and does not start any subshell (unlike kubie). As a result it works extremely fast

// TODO add little screencast

## Installation

Run

```shell
go install github.com/simontheleg/konf-go@latest
```

Afterwards, add the following to your `.zshrc` / `.bashrc` and restart your shell or re-source this file:

```zsh
# mandatory konf settings. This will install a shell wrapper called "konf" for you to use.
# Always use this wrapper, never call the konf-go binary directly!
# Currently supported shells: zsh, bash
source <(konf-go shellwrapper zsh)

# optional konf settings
# Alias
alias kctx="konf set"
alias kns="konf ns"
# Open last konf on new session (use --silent to suppress INFO log line)
export KUBECONFIG=$(konf-go --silent set -)
```

## Usage

Before any kubeconfig can be used with konf you have to import it:

```sh
konf import <path-to-your-kubeconf>
```

This is required, because konf maintains its own store of kubeconfigs to be able to work its "no-additional-shell-required"-magic.

Afterwards you can quickly switch between konfs using either:

```sh
konf set      # will open a picker dialogue
konf set -    # will open the last konf
konf set <id> # will set a specific konf. <id> is usually <context>_<cluster>
```

Additional commands and flags can be seen by calling `konf --help`

## How does it work?

### kubeconfig management across shells

Essentially konf maintains its state via two directories:

- `<konfDir>/store` -> contains all of your imported kubeconfigs, where each context is split into its own file
- `<konfDir>/active` -> contains all currently active konfs. The filename refers to the PID of the shell. Konf will automatically clean unused files after you close the session

We need these two extra directories because:

- each konf file must only contain one context. This is because konf can only use the `$KUBECONFIG` variable to point to one kubeconfig file. If there are multiple contexts in that file, kubernetes looks for a `current-context` key and sets the config to that, thus introducing some ambiguity. To avoid this, konf import splits all the contexts into separate files
- in order to allow for different shells to have different kubeconfigs we need to maintain a single one per shell. Otherwise when you run modifications like changing the namespace, these would affect all shells, which is not what we want

### zsh/bash-func-magic

One of the largest difficulties in this project lies in the core design of the shell.
Essentially a child process cannot make modifications to its parents.
This includes setting an environment variable, which affects us because we want to set `$KUBECONFIG`.
The way we work around this "limitation" is by using a zsh/bash function that executes our binary and then sets `$KUBECONFIG` to the output of `konf-go`.
With this trick we are able to set `$KUBECONFIG` and can make this project work. Since only the result of stdout will be captured by the zsh/bash-func, we can still communicate normally with the user by using stderr.

## Contributing

### Usage of stdout and stderr

When developing for konf, it is important to understand the konf-shellwrapper. It is designed to solve the problem described in the [zsh/bash-func-magic section](###zsh/bash-func-magic).
It works by saving the stdOut of konf-go in a separate variable and then evaluating the result. Should the result contain the keyword `KUBECONFIGCHANGE:`, the wrapper will set `$KUBECONFIG` to the value after the colon.
Otherwise the wrapper ist just going to print the result to stdOut in the terminal. This setup allows for konf-go commands to print to stdOut (which is required for example for zsh/bash completion). Additionally it should be able to handle large stdOut outputs as well, as it only parses the first line of output.
Interactive prompts however (like promptUI) should always print their dialogue to stdErr, as the wrapper has troubles with user input. Nonetheless you can still easily submit the result of the selection to the wrapper later on using the aforementioned keyword. So it should not be a big issue.

### Tests

By default `go test ./...` will run both unit and integration tests. Integration tests are mainly used to check for filename validity and only write in the `/tmp/konf` directory. They are mainly being used by the CI. If you only want to run unit-test, you can do so by using the `-short` flag:

```sh
go test -short ./...
```

If you want to only run integration tests, simply run:

```sh
go test -run Integration ./...
```

## Ideas for Future Improvements

- Make it work for fish
- Allow usage of other fuzzy finders like fzf
- `konf manage` so you can rename contexts and clusters
- `konf delete` option so you can delete konfs you don't need anymore
- figure out auto-completion. This might be a bit tricky due to the `konf` zsh func wrapper
- File column could be improved by either using '...' abreviation or filtering out the konfDir
