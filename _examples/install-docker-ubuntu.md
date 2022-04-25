# Install Docker on Ubuntu

Source: https://docs.docker.com/engine/install/ubuntu/ 

To install Docker Engine, you need the 64-bit version of one of these Ubuntu versions:

Ubuntu Impish 21.10
Ubuntu Hirsute 21.04
Ubuntu Focal 20.04 (LTS)
Ubuntu Bionic 18.04 (LTS)


## Uninstall old versions

Older versions of Docker were called docker, docker.io, or docker-engine. If these are installed, uninstall them:

```shell
sudo apt-get remove docker docker-engine docker.io containerd runc
```

## Install using the repository

Before you install Docker Engine for the first time on a new host machine, you need to set up the Docker repository. Afterward, you can install and update Docker from the repository.


### Setup this repository

* Update the apt package index and install packages to allow apt to use a repository over HTTPS:

```shell
sudo apt-get update
```

```shell
sudo apt-get install \
    ca-certificates \
    curl \
    gnupg \
    lsb-release
```

* Add Docker’s official GPG key:

```shell
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
```

* Use the following command to set up the stable repository.

```shell
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
```

### Install Docker Engine

* Update the apt package index, and install the latest version of Docker Engine and containerd, or go to the next step to install a specific version:

```shell
sudo apt-get update
```

```shell
sudo apt-get install docker-ce docker-ce-cli containerd.io
```

* Verify that Docker Engine is installed correctly by running the hello-world image.

```shell
sudo docker run hello-world
```


