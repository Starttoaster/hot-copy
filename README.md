# SyncAssist

## What this is

This is meant to run alongside a Syncthing instance. Syncthing is an impressive self-hosted distributed file syncing service that syncronizes your files between nodes. It encrypts your documents in transit, however, it lacks the functionality to keep data encrypted at rest. What this means is: when your data leaves or enters a node, it is encrypted, and then immediately unencrypted for the destination node.

The problem this repo is meant to solve is the ability to treat some nodes as untrustworthy relay servers. Say for instance, you have three Syncthing servers. Server A and C are only on while you're at work or home. Server B, in some public cloud server, is up at all times. You want server B to only have encrypted copies of your documents, but you want servers A and C to be able to pull the encrypted files, and encrypt/unencrypt on the fly.

## What this does

This repo takes a password from you, hashes it using SHA256 to make a 32 byte hash. That hash is used to encrypt all of your files using the AES-256 encryption standard. This is meant to run in a Docker container with two volumes. The `data/` volume is meant to face you. You store and access your unencrypted documents here. The `inside/` volume stores replicas of all of your files in encrypted form. That `inside/` volume can then be mounted to your Syncthing instance as its own data volume, so that it syncronizes only the encrypted files to the Syncthing relay server.

## TODO

A lot. This isn't done, by a long shot. For starters:

 - Add go unit tests
 - Write a better README.md file
 - Add Dockerfile to run in Docker