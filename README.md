# Use

 - Create a directory named "data" and another named "inside". The `data` directory stores your unencrypted data; the `inside` directory stores your encrypted data

 - Set your password with: `export SA_PASSWORD="my_super_secret_password"`

 - Build the binary: `go build`

 - Watch as data in your data in `data/` gets copied to `inside/`.

# TODO

A lot. This isn't done, by a long shot. For starters:

 - Make it so that if files in `data/` directory are deleted, they are also deleted from `inside/`, and vice versa
 - The above would require a check for if the removed file existed on the opposite directory during the previous run. If it did, then the file should delete. If it did not, then it should simply be copied (and {de/en}crypted)
 - Add go unit tests
 - Write a better README.md file
 - Add Dockerfile to run in Docker