![Build](https://github.com/1xyz/pryrite/workflows/Build/badge.svg)
![Release](https://github.com/1xyz/pryrite/workflows/Release/badge.svg)

# Pryrite

Pryrite is a command line tool that interactively runs executable blocks in a markdown file. One can think of pryrite as a console REPL/debugger for a markdown file. We drew inspiration from ruby's [pry](https://github.com/pry/pry) for the interface.

## Motivation

As developers, we come across plentiful documents that provide a prescriptive set of steps to do a task. Typically these include runbooks, onboarding/setup or trouble-shooting documents. Running commands from these documents involve: Reading the document in a browser followed by copy/pasting the commands in a console in order to run. We find this process cumbersome. Moreover, the output of the command(s) are never recorded, so you can never go back to a document and see if you had run a specific command, and what that result was.

## Features

Pryrite attempts to solve this by providing these features:

1. An interactive runner for markdown so that blocks marked as \`\`\`shell can be executed. 
	* (Note: we plan to support more executable types in the near future).
3. The result of each executable block is stored so that it can be retrieved contextually.


## Demo Example

The following video shows pryrite working with a markdown file.

https://user-images.githubusercontent.com/58613330/166069790-8de28d76-7103-40e0-bac3-f10b34df1655.mp4


* Try it out yourself

```shell
pryrite open https://raw.githubusercontent.com/1xyz/pryrite/main/_examples/hello-world.md
```


## Install

* MacOS

	```shell
	brew tap 1xyz/pryrite
	brew install pryrite
	```

* Linux (amd64)

	```shell
	wget -O pryrite https://github.com/1xyz/pryrite/releases/latest/download/pryrite-linux-amd64 && chmod 755 pryrite
	````

## Releases

Latest released binaries are available [here](https://github.com/1xyz/pryrite/releases/latest/). This includes other architectures.


## Examples to try

* Ubuntu

	Install docker on Ubuntu
	```shell
	pryrite open https://raw.githubusercontent.com/1xyz/pryrite/main/_examples/install-docker-ubuntu.md
	```

	Cleanup diskspace on Ubuntu

	```shell
	pryrite open https://raw.githubusercontent.com/1xyz/pryrite/main/_examples/cleanup-diskspace-ubuntu.md
	```

* MacOS

	Install and work with Redis
	```shell
	pryrite open https://raw.githubusercontent.com/1xyz/pryrite/main/_examples/install-redis-osx.md
	```



