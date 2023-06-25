# AWS ECR Registry cleaner

[![Go](https://github.com/manics/aws-ecr-registry-cleaner/actions/workflows/build.yml/badge.svg)](https://github.com/manics/aws-ecr-registry-cleaner/actions/workflows/build.yml)

A simple utility to delete ECR images older than a specified number of days.

This utility scans all repositories in an ECR registry and deletes images that haven't been pulled.

ECR supports [lifecycle policies](https://docs.aws.amazon.com/AmazonECR/latest/userguide/LifecyclePolicies.html), but currently deleting an image that hasn't been pulled for a specified number of days (`sinceImagePulled`) is not supported:
https://github.com/aws/containers-roadmap/issues/921

If it's added in the future this utility will not be needed.

## TODOs

- pagination for listing repositories and images
- continually rerun with a delay

## Build and run locally

You must install [Go 1.20](https://tip.golang.org/doc/go1.20).
If you are a Python developer using [Conda](https://docs.conda.io/en/latest/) or [Mamba](https://mamba.readthedocs.io/) and just want a quick way to install Go:

```
conda create -n go -c conda-forge go=1.20 go-cgo=1.20
conda activate go
```

```
make build
make test
```

Run with Amazon Web Services using the local [AWS credentials](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html):

## Build and run container

```
podman build -t aws-ecr-registry-cleaner .
```

Amazon Web Services:

```
podman run --rm -it \
  -eAWS_REGION=region \
  -eAWS_ACCESS_KEY_ID=access-key \
  -eAWS_SECRET_ACCESS_KEY=secret-key \
  -eAWS_ECR_EXPIRES_AFTER_PULL_DAYS=7 \
  aws-ecr-registry-cleaner
```

## Running in the cloud

The recommended way to run this service is to use an IAM
[instance profile (AWS)](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html)
to authenticate with the cloud provider.

### Environment variables

The following environment variables are supported:

- `AWS_ECR_EXPIRES_AFTER_PULL_DAYS`: Delete images which haven't been pulled for at least this number of days, default `7`.
- `AWS_REGISTRY_ID`: Registry ID to use for AWS ECR, only set this is you are not using the default registry for the AWS account.

## Development

Build and run

```
make build
make test
```

For more detailed testing of a single test:

```
go test -v ./amazon/ -run TestScanAndDeleteRepos
```

Update modules:

```
make update-deps
```
