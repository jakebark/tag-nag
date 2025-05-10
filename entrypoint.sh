#!/bin/sh
set -e 

terraform init -input=false -no-color || { echo "Terraform init failed"; exit 1; }

exec tag-nag "$@"
