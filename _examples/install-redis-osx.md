# Redis on MacOS

Source: https://gist.github.com/tomysmile/1b8a321e7c58499ef9f9441b2faa0aa8

This is an example markdown file showing some steps to work with  Redis on a Mac

* Update brew formulae and install redis via brew

## Install Redis via brew

```shell
brew update
brew install redis
```
## How to start/stop the service.

* To have launchd start redis now and restart at login:
```shell
brew services start redis
```

* to stop it, just run:
```shell
brew services stop redis
```

* Or, if you don't want/need a background service you can just run:
```shell
redis-server /usr/local/etc/redis.conf
```

## Test if Redis server is running.

```shell
redis-cli ping
```
If it replies “PONG”, then it’s good to go!

## Delve into the config

* Location of Redis configuration file: `/usr/local/etc/redis.conf`

```shell
export REDIS_CONFIG_FILE=/usr/local/etc/redis.conf
```

* You can edit inspect it 

```shell
vi $REDIS_CONFIG_FILE
```

## Uninstall Redis and its files.

```shell
brew uninstall redis
```
