# HotCopy

## What this is

Simply put, this is a docker image for bidirectional input and output between encrypted and unencrypted data. Using the SHA256 hashing algorithm for a user configured password, and the AES-256 encryption standard, this manages a directory of unencrypted data, and encrypts-decrypts it on the fly (preserving the original directory tree.)

The `/data` volume is meant to face the user with unencrypted data. The `/inside` volume contains the encrypted copies of your data.

#### What this is meant to solve

This is meant to run alongside something such as a Syncthing instance. Syncthing is an impressive self-hosted, distributed, file syncing service. For security reasons, it encrypts your documents in transit, however, it lacks the functionality to keep data encrypted at rest. What this means is, when your data exits or enters a node it is encrypted, and then immediately unencrypted for the destination node.

The problem this repo is meant to solve is the ability to treat some nodes as untrustworthy relay servers. If you have a cloud hosted instance of Syncthing, you may not want to allow it to store unencrypted copies of your personal data. 

To give a more detailed example, say that you have three Syncthing servers. Servers A and C are only on while you're either at work or home. Server B, in some public cloud server, is up at all times. You may want server B to only have encrypted copies of your documents, but you want servers A and C to be able to pull the encrypted files, and encrypt/decrypt on the fly. That is what this image is for. On servers A and C, you would simply need to mount this image's `/inside` volume to the data directory coming from your Syncthing instance, and you're done!

## Quick Start

Change the `SA_PASSWORD` environment variable to your desired password. Change the PUID/PGID environment variables to your user's values for best results. Find with `id <username>`

 - Build the image: `docker build -t hot-copy .`
 - Run the image: `docker container run --restart unless-stopped -v ~/path/to/data:/data -v ~/path/to/inside:/inside -e SA_PASSWORD="MyNewPassword" -e PUID="1000" -e PGID="1000" --name hot-copy hot-copy`

## TODO

This repo is still in early development. Feel free to contribute or send a pull request if you have an idea. For starters:

 - Further debugging, dealing with large numbers of files and directories, or large files
 - Further Go unit testing
 - Set up CI

 To test the image: `docker container run -v ~/go/src/hot-copy/data:/data -v ~/go/src/hot-copy/inside:/inside -e SA_PASSWORD="MyNewPassword" -e PUID="1000" -e PGID="1000" --rm --name hot-copy hot-copy`
