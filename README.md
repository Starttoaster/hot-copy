# SyncAssist

## What this is

This is meant to run alongside a Syncthing instance. Syncthing is an impressive self-hosted, distributed, file syncing service. For security reasons, it encrypts your documents in transit, however, it lacks the functionality to keep data encrypted at rest. What this means is, when your data exits or enters a node, it is encrypted, and then immediately unencrypted for the destination node.

The problem this repo is meant to solve is the ability to treat some nodes as untrustworthy relay servers. If you have a cloud hosted instance of Syncthing, you may not want to allow it to store unencrypted copies of your personal data. 

To give a more detailed example, say for instance that you have three Syncthing servers. Servers A and C are only on while you're either at work or home. Server B, in some public cloud server, is up at all times. You may want server B to only have encrypted copies of your documents, but you want servers A and C to be able to pull the encrypted files, and encrypt/unencrypt on the fly.

### What this does

This repo takes a password from you, hashes it using SHA256 to make a 32 byte hash. That hash is used to encrypt all of your files using the AES-256 encryption standard. This is meant to run in a Docker container with two volumes. The `data/` volume is meant to face you. You store and access your unencrypted documents here. The `inside/` volume stores replicas of all of your files in encrypted form. That `inside/` volume can then be mounted to your Syncthing instance as its own data volume, so that it syncronizes only the encrypted files to the Syncthing relay server. On the other end, using this same container in the same way, it will see that new versions of your files have arrived and unencrypt them for your second trusted node.

 ## Quick Start

 - Build the image: `docker build -t sync-assist .`
 - Run the image: `docker container run --restart unless-stopped -v ~/path/to/data:/data -v ~/path/to/inside:/inside -e SA_PASSWORD="MyNewPassword" --name sync-assist sync-assist`

## TODO

This repo is still in early development. Feel free to contribute or send a pull request if you have an idea. For starters:

 - Further testing dealing with large numbers of files and directories
 - Write Go unit tests
 - Set up CI

 To test the image: `docker container run -v ~/go/src/sync-assist/data:/data -v ~/go/src/sync-assist/inside:/inside -e SA_PASSWORD="MyNewPassword" -e PUID="1000" -e PGID="1000" --rm --name sync-assist sync-assist`
