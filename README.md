![Build](https://github.com/1xyz/pryrite/workflows/Build/badge.svg)
![Release](https://github.com/1xyz/pryrite/workflows/Release/badge.svg)

# Pryrite

Pryrite is an interactive runner for executable blocks defined in a markdown file.

In a way, you can think of pryrite as a REPL/debugger console for a markdown file. We drew inspiration from ruby's [pry](https://github.com/pry/pry) for the console interface.

## Problem

As a developer, we have come across plentiful documents that provide a prescriptive set of steps to do a task. Typically these include runbooks, onboarding/setup or trouble-shooting documents. Running commands from these documents involve: Reading the document in a browser followed by copy/pasting the commands in a console. We find this process cumbersome at the very best. Also the output of the command(s) are never recorded, so you could back to a document and see if you had ever run a specific command successfuly.

Pryrite provides the following:

1. Ability to load a markdown document in the terminal console so that executable blocks (blocks marked as \`\`\`shell) can be executed
2. The result of each executable block is stored as a history so that it can be retrieved contextually.

Here is a simple demo.


## Try it out

Examples:

1)

2)

