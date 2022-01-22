# Konf - Lightweight kubeconfig Manager

- [Konf - Lightweight kubeconfig Manager](#konf---lightweight-kubeconfig-manager)
  - [Why konf?](#why-konf)
  - [Installation](#installation)
  - [Usage](#usage)
  - [How does it work?](#how-does-it-work)
    - [kubeconfig management across shells](#kubeconfig-management-across-shells)
    - [zsh-func-magic](#zsh-func-magic)
  - [Contributing](#contributing)
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

Aferwards, add the following to your `.zshrc` and restart your shell or re-source the `.zshrc` afterwards:

```zsh
# mandatory konf settings
konf() {
  res=$(konf-go $@)
  # protect against an empty command
  # Note we cannot do something like if "$1 == set" and only run the export on set commands as cmd flags can be at any position in our cli
  if [[ $res != "" ]] then
    export KUBECONFIG=$res
  fi
}
konf_cleanup() {
  konf-go cleanup
}
add-zsh-hook zshexit konf_cleanup


# optional konf settings
# Alias
alias kctx="konf set"
alias kns="konf ns"
# Open last konf on new session
export KUBECONFIG=$(konf-go set -)
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

### zsh-func-magic

One of the largest difficulties in this project lies in the core design of the shell. Essentially a child process cannot make modifications to its parents. This includes setting an environment variable, which affects us because we want to set `$KUBECONFIG`. The way we work around this "limitation" is by using a zsh function that executes our binary and then sets `$KUBECONFIG` to the output of `konf-go`. With this trick we are actually able to set `$KUBECONFIG` and can make this project work. Since only the result of stdout will be captured by the zsh-func, we can still communicate normally with the user by using stderr.

As a result of this trick, konf currently only works with zsh. I did not have the time to see if I can also make it work for bash or fish. Maybe this becomes a community contribution? :)

## Contributing

When developing for konf, it is important to keep in mind that you can never print anything to stdout. You always have to use stderr. This is because anything printed to stdout will automatically be added to the `$KUBECONFIG` variable by the surrounding zsh-func.
Usually this should not be too much of an issue, because components like the default logger or promptui can both use stderr.

## Ideas for Future Improvements

- Make it work with other shells like bash or fish
- Maybe you can print zsh_func directly from a command?
- Allow usage of other fuzzy finders like fzf
- Add CI
- Double check the root.go and command descriptions
- `--silent` option for `set` command on which it does not log anything. This can be useful for things like `konf set -` when running it in every new session
- `konf manage` so you can rename contexts and clusters
- `konf delete` option so you can delete konfs you don't need anymore
- figure out auto-completion. This might be a bit tricky due to the `konf` zsh func wrapper
- File column could be improved by either using '...' abreviation or filtering out the konfDir
