[![Build](https://github.com/aardlabs/terminal-poc/workflows/Build/badge.svg)](https://github.com/aardlabs/terminal-poc/actions?query=workflow%3ABuild)
[![Release](https://github.com/aardlabs/terminal-poc/workflows/Release/badge.svg)](https://github.com/aardlabs/terminal-poc/actions?query=workflow%3ARelease)

## Pruney (CLI)

### Releases
- latest release of pruney from [here](https://github.com/aardlabs/terminal-poc/releases/latest)
- All releases are can be found at [releases](https://github.com/aardlabs/terminal-poc/releases/) page

### Installation notes

* For macOS - Allowing to run unverified developer program

```
# Unzip the downloaded binary in a separate folder
$ unzip pruney-darwin-amd64-v0.2.zip

# Run pruney --help
$ pruney --help

# On MacOS, you will get a message window indicating “pruney” cannot be opened because the developer ..
# .. cannot be verified.
# Go to System Preferences > Security & Privacy > General (tab)
# Under the General tab, Click on [Allow Anyway] for "pruney" was blocked ...

# ReRun pruney --help
$ pruney --help

# On MacOS, you will get a message window indicating "macOS cannot verify the developer of “pruney”...
# ..Are you sure you want to open it?"
# Click [Open]

# Phew!
```

* bash specific setup
    - Currently recording history is supported on bash shell.
    - The following allows pruney to record your history in bash

    ```
    # Run bash (if you are in another shell)
    $ bash

    # this will allow pruney to record your history
    bash-3.2$ source bash-init.sh

    # You can add the above command to your profile/configuration (i.e ~/.bashrc, ~/.profile, ~/.bash_profile, etc).
    # It must be the last thing imported in your bash profile. Basically don't add a PROMPT_COMMAND after this!
    ```

### Development notes
* Run `make`, this provides a list of options for development
* To create a release create and push a git tag
    ```
     $ git tag -a v0.2 -m "Release something"
     $ git push origin v0.2

    ```