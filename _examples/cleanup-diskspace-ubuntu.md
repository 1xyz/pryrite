# Ways to cleanup disk-space on ubuntu

Source: https://itsfoss.com/free-up-space-ubuntu-linux/

## Get rid of packages that are no longer required (Recommended)

This option removes libs and packages that were installed automatically to satisfy the dependencies of an installed package. If that package is removed, these automatically installed packages are useless in the system.

```shell
sudo apt-get autoremove
```

## Clean up APT cache in Ubuntu

The APT package management system keeps a cache of DEB packages in /var/cache/apt/archives. Over time, this cache can grow quite large and hold a lot of packages you don’t need.

* Check cache usage

```shell
sudo du -sh /var/cache/apt 
```

* Either remove only the outdated packages, like those superseded by a recent update, making them completely unnecessary.

```shell
sudo apt-get autoclean
```

* Or Delete the cache in its entirety

```shell
sudo apt-get clean
```

## Clear systemd journal logs

Check the log size with this command:

```shell
journalctl --disk-usage
```

The easiest for you is to clear the logs that are older than a certain days.

```shell
sudo journalctl --vacuum-time=3d
```

## Remove older versions of Snap applications

Snap packages are bigger in size. On top of that, Snap stores at least two older versions of the application (in case, you want to go back to the older version). This eats up huge chunk of space. In my case, it was over 5 GB.

```shell
du -h /var/lib/snapd/snaps
```

Alan Pope, part of Snapcraft team at Canonical, has created a small script that you can use and run to clean all the older versions of your snap apps

```shell
#!/bin/bash
# Removes old revisions of snaps
# CLOSE ALL SNAPS BEFORE RUNNING THIS
set -eu
sudo snap list --all | awk '/disabled/{print $1, $3}' |
    while read snapname revision; do
        sudo snap remove "$snapname" --revision="$revision"
    done
```




