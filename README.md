[![Build](https://github.com/aardlabs/terminal-poc/workflows/Build/badge.svg)](https://github.com/aardlabs/terminal-poc/actions?query=workflow%3ABuild)
[![Release](https://github.com/aardlabs/terminal-poc/workflows/Release/badge.svg)](https://github.com/aardlabs/terminal-poc/actions?query=workflow%3ARelease)

## aard (CLI)

### Releases
- Latest release of the CLI from [here](https://github.com/aardlabs/terminal-poc/releases/latest)
- All releases are can be found at [releases](https://github.com/aardlabs/terminal-poc/releases/) page
- All releases are mirrored [here](https://github.com/aardlabs/cli-release), so that they can be accessed publicly

### Development notes
* Run `make`, this provides a list of options for development
* To create a release update version.txt, commit and push to main. This triggers a release github action.

### Release a staging build
* A staging build is a pre-release build used for testing. Here are the steps to do this:
  - Update the branch `staging` from `origin/main` and push to remote (`origin/staging`)
  - Go to the release github action [here](https://github.com/aardlabs/terminal-poc/actions/workflows/release.yml)
    - Build the release via the `Run workflow` button; choosing `branch = staging`