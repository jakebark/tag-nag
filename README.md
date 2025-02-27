# tag nag

Validate AWS tags in terraform.

Designed to be part of pre-deployment checks, ensuring compliance with your tagging strategy. 

## Installation
```bash
go install github.com/jakebark/tag-nag@latest
```
You may need to set [GOPATH](https://go.dev/wiki/SettingGOPATH)

### Docker
```bash
docker pull jakebark/tag-nag:latest
docker run --rm -v $(pwd):/workspace jakebark/tag-nag --tags "Owner,Environment" /workspace

```

## Commands

Tag nag will search a file or directory for tag keys. 

```bash
tag-nag <file/directory> --tags "<tagKey1>,<tagKeyN>"
-c # case-insensitive 

tag-nag foo.tf --tags "Owner" # run against a file
tag-nag ./bar --tags "Owner","Environment" # run against a directory

```
