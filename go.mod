module github.com/manics/aws-ecr-registry-cleaner

go 1.20

require (
	github.com/aws/aws-sdk-go-v2 v1.23.1
	github.com/aws/aws-sdk-go-v2/config v1.25.4
	github.com/aws/aws-sdk-go-v2/service/ecr v1.23.1
	github.com/aws/aws-sdk-go-v2/service/sts v1.25.4
)

require (
	github.com/aws/aws-sdk-go-v2/credentials v1.16.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.17.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.20.1 // indirect
	github.com/aws/smithy-go v1.17.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)

replace github.com/manics/aws-ecr-registry-cleaner/amazon => ./amazon
