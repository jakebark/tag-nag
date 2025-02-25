# tag nag

Validate AWS tags in terraform.

Designed to be part of pre-deployment checks, ensuring compliance with your tagging strategy. 

## Installation
```bash
go install github.com/jakebark/tag-nag@latest
```

### Docker
```bash
docker pull jakebark/tag-nag:latest
docker run --rm -v $(pwd):/workspace jakebark/tag-nag --tags "Owner,Environment" /workspace

```

## Commands

```bash
tag-nag <file/directory> --tags "<tagKey1>,<tagKey2>"

tag-nag foo.tf --tags "bar"
tag-nag ./foo --tags "bar","baz"

```
