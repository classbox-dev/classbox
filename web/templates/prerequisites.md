{{define "title"}}Prerequisites @ hsecode{{end -}}
# Prerequisites

{{if not .User -}}
[Sign in](signin) to open the page.
{{- else -}}

Why, sure, you *can* write and submit code using that silly editor on GitHub. However, *efficient* programming requires some proper development tools.

**Note for Windows users:** (tl;dr do not use WSL.) As with most open-source tools, Go is much better supported on Unix-like systems. This may tempt you to use [Windows Subsystem for Linux](https://docs.microsoft.com/en-us/windows/wsl/about) (WSL), which is a fully-fledged Ubuntu distribution inside a running Windows. However, it is still somewhat hard to integrate the Windows-world (filesystem and graphical IDE) with the WSL-world. Unless you want to write your code in Vim or deal with all sorts of weird hacks, I do not recommend using WSL (yet). This tutorial goes on with installing Windows-native tools. As far as I can tell, for this course they work just fine.

## Install Go

You are expected to have Go version 1.13.x on your computer.

### Linux, macOS

* You are better off installing it [from a .tar.gz archive](https://golang.org/doc/install#tarball), **not** with a package manager (unless you know what you are doing).
* Make sure to add {{"`<...>/go/bin`" |unescape}}  to the PATH environment variable as described in the instruction.
* Set up a directory for go packages, I recommend `$HOME/.local/go`:
  ```
  mkdir -p $HOME/.local/go
  go env -w GOPATH=$HOME/.local/go
  ```
  Also add `$HOME/.local/go/bin` to the PATH.
* Log off and log on again to apply the changes to PATH
* Run `go env -w GO111MODULE=on`

### Windows

* I recommend using the [MSI installer](https://golang.org/doc/install#windows). It will automatically set all the environment variables for you.
* If you installed Go somehow else, ensure that you can run `go version` either in command prompt (`Win + r` then type `cmd`) or a powershell window (`Win + i`). Otherwise, find the location of `go.exe` (typically `C:\Go\bin`) and insert it to the `PATH`:
  ```
  setx PATH "%PATH%;C:\Go\bin"
  ```
  Then run the following commands to set up a directory for go packages:
  ```
  md %USERPROFILE%\go
  setx GOPATH %USERPROFILE%\go
  setx PATH "%PATH%;%USERPROFILE%\go\bin"
  ```

In any case, run `go env -w GO111MODULE=on`

### Test Your Installation

You should be able to run the following command in the terminal:
```
$ go version
go version go1.13.7 linux/amd64

$ gofmt --help
usage: gofmt [flags] [path ...]
...

$ go env GO111MODULE
on
```

## Install Git

I strongly recommend using Git as a command-line tool, **not** as a GUI app (such as GitHub Desktop/SourceTree/etc).

* Linux/macOS: you probably already have git installed.
* Windows: unless you already have some collection of Linux-like tools (MinGW, Cygwin, or their distributions), install [Git for Windows](https://gitforwindows.org). It comes with Bash ("Git Bash"), GNU tools, and a separate terminal emulator (which you may or may not want to use).

Check the installation:

```
$ git --version
git version 2.17.1
```

## (optional) Setup SSH Keys

Your working repository is private. If you clone it via HTTPS (`git clone https://...`), you will have to enter your GitHub login and password on for each operation. It is rather annoying. Instead, I recommend cloning via SSH (`git clone git@github.com:...`). It requires setting up a pair of SSH keys: for your computer (private key) and the repository (public key). Once you do that, you will not have to supply passwords for git ever again.

Follow the instructions on GitHub:

1. [Generating a new SSH key and adding it to the ssh-agent](https://help.github.com/en/github/authenticating-to-github/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent)
2. [Adding a new SSH key to your GitHub account](https://help.github.com/en/github/authenticating-to-github/adding-a-new-ssh-key-to-your-github-account)

## Clone Working Repository

Clone the [working repository](https://github.com/{{.User.Login}}/{{.User.Repo}}) somewhere on your computer:

```
$ git clone git@github.com:{{.User.Login}}/{{.User.Repo}}.git
```

OR use https URL if you decided not to install SSH keys:

```
$ git clone https://github.com/{{.User.Login}}/{{.User.Repo}}.git
```


## Install genny

You will also need [genny](https://github.com/cheekybits/genny), code-generator for Go. More on that in [Quickstart](../quickstart).
```
$ go install github.com/cheekybits/genny
go: finding github.com/cheekybits/genny v1.0.0
go: downloading github.com/cheekybits/genny v1.0.0
...
```

Make sure it is successfully installed:
```
$ genny
usage: genny [{flags}] gen "{types}"
...
```

## (optional) Install pre-commit

[pre-commit](https://pre-commit.com) is a tool that performs predefined checks on your code before every commit. In some cases, it will not let you commit before you fix certain issues.

In out case pre-commit performs the following actions:

* Runs `go generate` (see [Quickstart](../quickstart) for details).
* Automatically fixes code-style.
* Runs [stdlib-linter](https://github.com/mkuznets/stdlib-linter) to ensure your code meets syntactic requirements, e.g. that you only use allowed packages.
* Ensures there are no untracked files in your repository.

It is optional: the test system will let you know about any of these issues anyway. However, you are better off to catch such things as early as possible, without waiting for your commits to be tested.

Follow the [installation instruction](https://pre-commit.com/#install) and make sure you can run pre-commit in a terminal:
```
$ pre-commit --version
pre-commit 2.0.1
```

Then install the hooks in your working repository:
```
$ cd hsecode-stdlib
$ pre-commit install
pre-commit installed at .git/hooks/pre-commit
```

## Install IDE

Skip this section entirely if you already have strong editor/IDE preferences. Just make sure your thing have at least some amount of Go support.

### Option 1: GoLand

As of February 2020, [GoLand](https://www.jetbrains.com/go/) is by far the best choice for programming in Go. It comes [free of charge for students](https://www.jetbrains.com/student/).

However, I understand if you do not want to deal with another big honking product from JetBrains. If that is the case, try VS Code as an alternative.

### Option 2: Visual Studio Code

[VS Code](https://code.visualstudio.com) has [officially bad](https://github.com/Microsoft/vscode-go/wiki/Go-modules-support-in-Visual-Studio-Code) support of Go modules. Still, as far as I can tell it has *just enough* of it for this course.

* [Install VS Code](https://code.visualstudio.com/Download)
* Install Go extension, either via *File -> Preferences -> Extensions* or from the terminal:
  ```
  $ code --install-extension ms-vscode.Go
  ```
* Open *View -> Command Palette*, type "Go: Install/Update Tools". Select all the items and press OK. It may end up with some failures, ignore them for now.
* Also in *Command Palette*, select "Preferences: Open Settings (JSON)" and paste the following:
  ```json
  {
    "go.useLanguageServer": true,
    "gopls": {
      "staticcheck": true,
      "completeUnimported": true,
      "deepCompletion": true,
      "usePlaceholders": true
    },
    "go.languageServerExperimentalFeatures": {
      "format": true,
      "autoComplete": true,
      "documentLink": false
    }
  }
  ```
* VS Code may continue to harass you about restarting or installing more things. Let it do whatever it says.
* In case of any problems read more about configuring the Go Language Server: [[1]](https://github.com/microsoft/vscode-go#go-language-server) [[2]](https://github.com/golang/tools/blob/master/gopls/doc/vscode.md)

Now hopefully you can open the `hsecode-stdlib` folder with *File -> Open folder* and everything will be just fine.

## Quickstart

Continue with [Quickstart](../quickstart) to start implementing stdlib.

{{end}}

* [Back to main page](..)
