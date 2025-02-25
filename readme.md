# tag nag

Validate AWS tags in terraform.

Designed to be part of pre-deployment checks, ensuring compliance with your tagging strategy. 

## Installation
```bash
go install github.com/jakebark/tag-nag@latest
```

## Commands

```bash
tag-nag <file/directory> --tags "<tagKey1>,<tagKey2>"

tag-nag foo.tf --tags "bar"
tag-nag ./foo --tags "bar","baz"

```
