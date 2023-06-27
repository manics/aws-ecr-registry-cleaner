# AWS ECR Registry cleaner

[![Go](https://github.com/manics/aws-ecr-registry-cleaner/actions/workflows/build.yml/badge.svg)](https://github.com/manics/aws-ecr-registry-cleaner/actions/workflows/build.yml)

A simple utility to delete ECR images older than a specified number of days.

This utility scans all repositories in an ECR registry and deletes images that haven't been pulled.

ECR supports [lifecycle policies](https://docs.aws.amazon.com/AmazonECR/latest/userguide/LifecyclePolicies.html), but currently deleting an image that hasn't been pulled for a specified number of days (`sinceImagePulled`) is not supported:
https://github.com/aws/containers-roadmap/issues/921

If it's added in the future this utility will not be needed.

## TODOs

- pagination for listing repositories and images

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

Delete images that haven't been pulled for at least 7 days, then exit:

```
aws-ecr-registry-cleaner -expires-after-pull-days=7
```

Delete all images

```
aws-ecr-registry-cleaner -expires-after-pull-days=0
```

Delete images that haven't been pulled for at least 7 days, then re-run after one hour:

```
aws-ecr-registry-cleaner -expires-after-pull-days=7 -loop-delay=3600
```

For for help run

```
aws-ecr-registry-cleaner -help
```

## Build and run container

```
podman build -t aws-ecr-registry-cleaner .
```

Delete images that haven't been pulled for at least 7 days:

```
podman run --rm -it \
  -eAWS_REGION=region \
  -eAWS_ACCESS_KEY_ID=access-key \
  -eAWS_SECRET_ACCESS_KEY=secret-key \
  aws-ecr-registry-cleaner -expires-after-pull-days=7
```

## Running in the cloud

The recommended way to run this service is to use an IAM
[instance profile (AWS)](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html)
to authenticate with the cloud provider.

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
