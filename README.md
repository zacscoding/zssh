# zssh

Simple ssh cli for persistent hosts information.  

![example](https://user-images.githubusercontent.com/25560203/146677584-c5127df7-3613-4e74-9095-96b170e77dee.gif)

# Getting started  

## (1) Go get and install

```shell
$ go get -u github.com/zacscoding/zssh/...
$ zssh --help
```

### (2) Go install with git clone

```shell
$ mkdir -p $GOPATH/src/github.com/zacscoding
$ cd $GOPATH/src/github.com/zacscoding
$ git clone https://github.com/zacscoding/zssh.git
$ cd zssh/cmd/zssh
$ go install
```

Check **zssh** in $GOPATH

```shell
$ ls -la $GOBIN/zssh
-rwxr-xr-x  1 evan.kim  staff  13667000 12 19 22:27 .../zssh
```

Don't forget to add `$GOPATH/bin` to ur `$PATH`.

# Usage

## Host commands  

```shell
$ zssh host --help
Handle host info

Usage:
  zssh host [command]

Available Commands:
  active      Get active a host
  add         Adds a new host info
  delete      Delete the host
  get         Get a host
  gets        Get host all
  select      Select a default host
  update      Update the host

Flags:
  -h, --help   help for host

Global Flags:
      --workspace string   workspace path(default: $HOME/.zssh)

Use "zssh host [command] --help" for more information about a command.
```  

## SSH Commands

```shell
$ zssh ssh --help
Commands for ssh

Usage:
  zssh ssh [command]

Available Commands:
  exec        Execute command to the remote host
  shell       Open the remote shell

Flags:
  -h, --help   help for ssh

Global Flags:
      --workspace string   workspace path(default: $HOME/.zssh)

Use "zssh ssh [command] --help" for more information about a command.
```