[![Build](https://github.com/aardlabs/terminal-poc/workflows/Build/badge.svg)](https://github.com/aardlabs/terminal-poc/actions?query=workflow%3ABuild)
[![Release](https://github.com/aardlabs/terminal-poc/workflows/Release/badge.svg)](https://github.com/aardlabs/terminal-poc/actions?query=workflow%3ARelease)

## aard (CLI)

### Releases
- latest release of the CLI from [here](https://github.com/aardlabs/terminal-poc/releases/latest)
- All releases are can be found at [releases](https://github.com/aardlabs/terminal-poc/releases/) page

### Installation notes

* For macOS
```
brew tap aardlabs/cli
brew install aard
```

* bash specific setup
    - Currently recording history is supported on bash shell.
    - The following allows the CLI to record your history in bash

    ```
    # Run bash (if you are in another shell)
    $ bash

    # this will allow the CLI to record your history
    bash-3.2$ source bash-init.sh

    # You can add the above command to your profile/configuration (i.e ~/.bashrc, ~/.profile, ~/.bash_profile, etc).
    # It must be the last thing imported in your bash profile. Basically don't add a PROMPT_COMMAND after this!
    ```

### Development notes
* Run `make`, this provides a list of options for development
* To create a release create and push a git tag
    ```
     $ git tag -a v0.2 -m "Release something"
     $ git push origin --follow-tags
    ```