# HotCopy

[![pipeline status](https://gitlab.com/brandonbutler/hot-copy/badges/master/pipeline.svg)](https://gitlab.com/brandonbutler/hot-copy/commits/master)
[![coverage report](https://gitlab.com/brandonbutler/hot-copy/badges/master/coverage.svg)](https://gitlab.com/brandonbutler/hot-copy/commits/master)

## What this is

Simply put, this is a docker image for bidirectional input and output between encrypted and unencrypted data. Using the SHA256 hashing algorithm for a user configured password, and the AES-256 encryption standard, this manages a directory of unencrypted data, and encrypts-decrypts it on the fly (preserving the original directory tree.)

This is meant to run alongside something such as a Syncthing instance. Syncthing is an impressive self-hosted, distributed, file syncing service. For security reasons, it encrypts your documents in transit, however, it lacks the functionality to keep data encrypted at rest. What this means is, when your data exits or enters a node it is encrypted, and then immediately unencrypted for the destination node.

### What this is meant to solve

The problem this repo is meant to solve is the ability to treat some nodes as untrustworthy relay servers. If you have a server hosted in a public cloud, you may not want to allow it to store unencrypted copies of your personal data. To give a more detailed example, see the diagram shared below with three computers all running Syncthing.

![Diagram](https://i.imgur.com/3HxjUh8.png)

## Quick Start

 - Build the image: `docker build -t hot-copy -f 'release/Dockerfile' .`
 - Run the image: 
 ```
 docker container run \
    --restart unless-stopped \
    -v ~/path/to/data:/data \
    -v ~/path/to/enc-data:/enc-data \
    -e SA_PASSWORD="MyNewPassword" \
    -e PUID="1000" \
    -e PGID="1000" \
    --name hot-copy \
    hot-copy
 ```

 ### Environment Variables and Volumes

| Variable/Volume | Function |
| ---- | ---- |
| -e SA_PASSWORD | This is the password you will use to encrypt and decrypt files with |
| -e PUID | Your host username's ID number. Find with `id <username>` |
| -e PGID | Your host username's group number. Find with `id <username>` | 
| -v /data | Your directory with unencrypted data |
| -v /enc-data | The encrypted copies of all of your data |


## TODO

This repo is still in early development. Feel free to contribute or send a pull request if you have an idea. For starters:

 - Further debugging, dealing with large numbers of files and directories, or large files
 - Further Go unit testing

#### For testing
```
docker build -t hot-copy-test -f 'test/Dockerfile' .
docker container run --rm hot-copy-test
```