<div align="center">
<img alt="tag:nag" height="200" src="./img/tag:nag.png" />
</div>

# tag:nag

<img src="./img/demo.gif" width="650">

Validate AWS tags in Terraform and CloudFormation.  

Designed to run in a pipeline or as part of pre-deployment checks.  

## Installation
```bash
go install github.com/jakebark/tag-nag@latest
```
You may need to set [GOPATH](https://go.dev/wiki/SettingGOPATH).

### Docker
```bash
docker pull jakebark/tag-nag:latest
docker run --rm -v $(pwd):/workspace jakebark/tag-nag --tags "Owner,Environment" /workspace

```

## Commands

Tag nag will search a file or directory for tag keys. 

```bash
tag-nag <file/directory> --tags "<tagKey1>,<tagKeyN>"

tag-nag main.tf --tags "Owner" # run against a file
tag-nag ./my_project --tags "Owner","Environment" # run against a directory

```

Search for tag keys *and* values:

```bash
tag-nag <file/directory> --tags "<tagKey1>=<tagValue1>"

tag-nag main.tf --tags "Owner=Jake" 
tag-nag main.tf --tags "Owner=Jake","Environment" # mixed search possible

```

Optional flags: 
```bash
-c # case-insensitive 
```
## Related Resources

- [pkg.go.dev/github.com/jakebark/tag-nag](https://pkg.go.dev/github.com/jakebark/tag-nag)
