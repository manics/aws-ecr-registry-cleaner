// https://aws.github.io/aws-sdk-go-v2/docs/
// https://aws.github.io/aws-sdk-go-v2/docs/handling-errors/

package amazon

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type IEcrClient interface {
	DescribeRepositories(ctx context.Context, input *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (response *ecr.DescribeRepositoriesOutput, err error)

	DescribeImages(ctx context.Context, input *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (response *ecr.DescribeImagesOutput, err error)

	BatchDeleteImage(ctx context.Context, input *ecr.BatchDeleteImageInput, optFns ...func(*ecr.Options)) (response *ecr.BatchDeleteImageOutput, err error)

	DeleteRepository(ctx context.Context, input *ecr.DeleteRepositoryInput, optFns ...func(*ecr.Options)) (response *ecr.DeleteRepositoryOutput, err error)

	DeleteLifecyclePolicy(ctx context.Context, input *ecr.DeleteLifecyclePolicyInput, optFns ...func(*ecr.Options)) (response *ecr.DeleteLifecyclePolicyOutput, err error)
}

type ecrDeletionHandler struct {
	registryId           string
	expiresAfterPullDays int
	client               IEcrClient
	dryRun               bool
}

func (c *ecrDeletionHandler) ScanAndDeleteRepos(olderThan *time.Time) []error {
	errs := []error{}

	input := ecr.DescribeRepositoriesInput{}
	if c.registryId != "" {
		input.RegistryId = &c.registryId
	}
	repos, err := c.client.DescribeRepositories(context.TODO(), &input)
	if err != nil {
		log.Println("ERROR:", err)
		errs = append(errs, err)
		return errs
	}

	for _, repo := range repos.Repositories {
		err := c.ScanAndDeleteImages(*repo.RepositoryName, olderThan)
		if err != nil {
			log.Println("ERROR:", err)
			errs = append(errs, err)
		}
	}

	return errs
}

func (c *ecrDeletionHandler) ScanAndDeleteImages(repoName string, olderThan *time.Time) error {
	empty := true
	input := ecr.DescribeImagesInput{
		RepositoryName: &repoName,
	}
	if c.registryId != "" {
		input.RegistryId = &c.registryId
	}
	images, err := c.client.DescribeImages(context.TODO(), &input)
	if err != nil {
		log.Println("ERROR:", err)
		return err
	}

	tagsToDelete := []string{}
	digestsToDelete := []string{}

	for _, image := range images.ImageDetails {
		lastUse := image.LastRecordedPullTime
		if lastUse == nil {
			log.Printf("WARN: Image %s [%v] has never been pulled, using last push\n", repoName, image.ImageTags)
			lastUse = image.ImagePushedAt
		}
		if lastUse.Before(*olderThan) {
			tagsToDelete = append(tagsToDelete, image.ImageTags...)
			digestsToDelete = append(digestsToDelete, *image.ImageDigest)
		} else {
			empty = false
		}
	}

	if len(digestsToDelete) > 0 {
		log.Printf("Deleting image %s %v\n", repoName, tagsToDelete)
		err := c.DeleteImage(repoName, &digestsToDelete)
		if err != nil {
			log.Println("ERROR:", err)
			return err
		}
	}

	if empty {
		err := c.DeleteRepository(repoName)
		if err != nil {
			log.Println("ERROR:", err)
			return err
		}
	}
	return nil
}

func (c *ecrDeletionHandler) DeleteImage(repoName string, imageDigests *[]string) error {
	if c.dryRun {
		log.Printf("DRYRUN: DeleteImage(%s, %v)", repoName, *imageDigests)
		return nil
	}

	input := ecr.BatchDeleteImageInput{
		ImageIds:       make([]types.ImageIdentifier, len(*imageDigests)),
		RepositoryName: &repoName,
	}
	if c.registryId != "" {
		input.RegistryId = &c.registryId
	}

	for n, digest := range *imageDigests {
		input.ImageIds[n].ImageDigest = aws.String(digest)
	}
	_, err := c.client.BatchDeleteImage(context.TODO(), &input)
	if err != nil {
		log.Println("ERROR:", err)
		return err
	}
	return nil
}

func (c *ecrDeletionHandler) deleteRepositoryPolicy(repoName string) error {
	input := ecr.DeleteLifecyclePolicyInput{
		RepositoryName: &repoName,
	}
	if c.registryId != "" {
		input.RegistryId = &c.registryId
	}
	_, err := c.client.DeleteLifecyclePolicy(context.TODO(), &input)
	if err != nil {
		// Ignore if it didn't exist
		var awsErrRepo *types.RepositoryNotFoundException
		var awsErrPolicy *types.LifecyclePolicyNotFoundException
		if errors.As(err, &awsErrRepo) || errors.As(err, &awsErrPolicy) {
			log.Println("Lifecycle policy not found", repoName)
			return nil
		}
		return err
	}
	log.Printf("Policy for repo '%s' deleted", repoName)
	return nil
}

func (c *ecrDeletionHandler) DeleteRepository(repoName string) error {
	if c.dryRun {
		log.Printf("DRYRUN: DeleteRepository(%s)", repoName)
		return nil
	}
	log.Println("Deleting repo", repoName)

	err := c.deleteRepositoryPolicy(repoName)
	if err != nil {
		log.Println("ERROR:", err)
		return err
	}

	input := ecr.DeleteRepositoryInput{
		RepositoryName: &repoName,
	}
	if c.registryId != "" {
		input.RegistryId = &c.registryId
	}

	_, err = c.client.DeleteRepository(context.TODO(), &input)
	if err != nil {
		log.Println("ERROR:", err)
		return err
	}
	return nil
}

func (c *ecrDeletionHandler) RunOnce() []error {
	olderThan := time.Now().Add(-time.Duration(c.expiresAfterPullDays) * 24 * time.Hour)
	log.Printf("Deleting images older than %s\n", olderThan.Format(time.RFC3339))
	errs := c.ScanAndDeleteRepos(&olderThan)
	return errs
}

func Setup(dryRun bool, registryId string, expiresAfterPullDays int) (*ecrDeletionHandler, error) {
	// Automatically looks for a usable configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("failed to load configuration, %v", err)
		return nil, err
	}

	stsClient := sts.NewFromConfig(cfg)
	identity, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Printf("failed to get identity, %v", err)
		return nil, err
	}
	log.Printf("Identity: %v", *identity.Arn)

	ecrClient := ecr.NewFromConfig(cfg)

	log.Println("Registry ID:", registryId)

	ecrH := &ecrDeletionHandler{
		registryId:           registryId,
		expiresAfterPullDays: expiresAfterPullDays,
		client:               ecrClient,
		dryRun:               dryRun,
	}

	return ecrH, nil
}
