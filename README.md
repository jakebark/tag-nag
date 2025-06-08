# tag-nag

<img src="./img/demo.gif" width="650">

Validate AWS tags in Terraform and CloudFormation.  

## Installation
```bash
go install github.com/jakebark/tag-nag@latest
```
You may need to set [GOPATH](https://go.dev/wiki/SettingGOPATH).

## Commands

Tag-nag will search a file or directory for tag keys. Directory search is recursive.

```bash
tag-nag <file/directory> --tags "Key1,Key2"

tag-nag main.tf --tags "Owner" # run against a file
tag-nag ./my_project --tags "Owner,Environment" # run against a directory
tag-nag . --tags "Owner", "Environment" # will take string or list

```

Search for tag keys *and* values

```bash
tag-nag <file/directory> --tags "Key[Value]"

tag-nag main.tf --tags "Owner[Jake]" 
tag-nag main.tf --tags "Owner[Jake],Environment" # mixed search possible
tag-nag main.tf --tags "Owner[Jake],Environment[Dev,Prod]" # multiple options for tag values

```

Optional flags 
```bash
-c # case-insensitive 
-d # dry-run (will always exit successfully)
```

Optional inputs
```bash
--cfn-spec ~/path/to/CloudFormationResourceSpecification.json # path to Cfn spec file, filters taggable resources
```

## Skip Checks
Skip file
```hcl
#tag-nag ignore-all
```

Terraform
```hcl
resource "aws_s3_bucket" "this" {
  #tag-nag ignore
  bucket   = "that"
}
```

CloudFormation
```yaml
EC2Instance:  #tag-nag ignore
    Type: "AWS::EC2::Instance"
    Properties: 
      ImageId: ami-12a34b
      InstanceType: c1.xlarge   
```

## Filtering taggable resources

Some AWS resources cannot be tagged. 

To filter out these resources with Terraform, run tag-nag against an initialised directory (`terraform init`).

To filter out these resources with CloudFormation, specify a path to the [CloudFormation JSON spec file](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cfn-resource-specification.html) with the `--cfn-spec` input. 

## Docker
Run
```bash
docker pull jakebark/tag-nag:latest
docker run --rm -v $(pwd):/workspace -w /workspace jakebark/tag-nag \
  . --tags "Owner,Environment" 

```

Interactive shell
```bash
docker pull jakebark/tag-nag:latest
docker run -it --rm \
  -v "$(pwd)":/workspace \
  -w /workspace \
  --entrypoint /bin/sh jakebark/tag-nag:latest
```

The image contains terraform, allowing `terraform init` to be run, if required.  
```bash
docker pull jakebark/tag-nag:latest
docker run --rm -v $(pwd):/workspace -w /workspace \
  --entrypoint /bin/sh jakebark/tag-nag:latest \
  -c "terraform init -input=false -no-color && tag-nag\
     . --tags 'Owner,Environment'"
```

## CI/CD

Example CI files:
- [GitHub](./examples/github.yml)
- [GitLab](./examples/gitlab.yml)
- [AWS CodeBuild](./examples/codebuild.yml)

## Related Resources

- [pkg.go.dev/github.com/jakebark/tag-nag](https://pkg.go.dev/github.com/jakebark/tag-nag)

<div align="center">
<img alt="tag:nag" height="150" src="./img/tag.png" />
</div>
