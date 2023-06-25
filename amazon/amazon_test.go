package amazon

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

type MockEcrClient struct {
	describeRepoRequests         []ecr.DescribeRepositoriesInput
	describeImageRequests        []ecr.DescribeImagesInput
	batchDeleteImageRepoRequests []ecr.BatchDeleteImageInput
	deleteRepoRequests           []ecr.DeleteRepositoryInput
	deleteLifecycleRequests      []ecr.DeleteLifecyclePolicyInput
	t                            *testing.T
}

const registryId = "123456789012"

func olderThanTime() *time.Time {
	return aws.Time(time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC))
}

func (c *MockEcrClient) repository(name string) types.Repository {
	return types.Repository{
		RegistryId:     aws.String(registryId),
		RepositoryName: &name,
		RepositoryUri:  aws.String(fmt.Sprintf("%s.dkr.ecr.eu-west-2.amazonaws.com/%s", registryId, name)),
	}
}

func (c *MockEcrClient) image(name string, tag string, lastPulled *time.Time) types.ImageDetail {
	digest := "digest-" + tag
	return types.ImageDetail{
		ImageTags:            []string{tag},
		ImageDigest:          &digest,
		RegistryId:           aws.String(registryId),
		RepositoryName:       &name,
		LastRecordedPullTime: lastPulled,
	}
}

func (c *MockEcrClient) DescribeRepositories(ctx context.Context, input *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (response *ecr.DescribeRepositoriesOutput, err error) {
	c.describeRepoRequests = append(c.describeRepoRequests, *input)

	if input.RepositoryNames != nil {
		c.t.Fatalf("Expected input.RepositoryNames to be empty")
	}
	return &ecr.DescribeRepositoriesOutput{
		Repositories: []types.Repository{
			c.repository("repo1"),
			c.repository("repo2"),
			c.repository("repo3"),
		},
	}, nil
}

func (c *MockEcrClient) DescribeImages(ctx context.Context, input *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (response *ecr.DescribeImagesOutput, err error) {
	c.describeImageRequests = append(c.describeImageRequests, *input)
	for _, imageId := range input.ImageIds {
		if imageId.ImageTag != nil || imageId.ImageDigest != nil {
			c.t.Fatalf("Expected ImageTag and ImageDigest to be empty")
		}
	}

	if *input.RepositoryName == "repo1" {
		return &ecr.DescribeImagesOutput{
			ImageDetails: []types.ImageDetail{
				c.image("repo1", "tag1", aws.Time(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC))),
				c.image("repo1", "tag2", aws.Time(time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC))),
			},
		}, nil
	}
	if *input.RepositoryName == "repo2" {
		return &ecr.DescribeImagesOutput{
			ImageDetails: []types.ImageDetail{
				c.image("repo2", "tag3", aws.Time(time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC))),
				c.image("repo2", "tag4", aws.Time(time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC))),
			},
		}, nil
	}
	if *input.RepositoryName == "repo3" {
		return &ecr.DescribeImagesOutput{}, nil
	}

	return nil, &types.ImageNotFoundException{Message: aws.String("Image not found")}
}

func (c *MockEcrClient) assertImageDigests(expected *[]string, actual *[]types.ImageIdentifier) {
	digests := []string{}
	for _, imageId := range *actual {
		if imageId.ImageTag != nil {
			c.t.Errorf("Expected ImageTag to be empty")
		}
		digests = append(digests, *imageId.ImageDigest)
	}

	if !reflect.DeepEqual(digests, *expected) {
		c.t.Errorf("Expected digests %s, got %s", *expected, digests)
	}
}

func (c *MockEcrClient) BatchDeleteImage(ctx context.Context, input *ecr.BatchDeleteImageInput, optFns ...func(*ecr.Options)) (response *ecr.BatchDeleteImageOutput, err error) {
	c.batchDeleteImageRepoRequests = append(c.batchDeleteImageRepoRequests, *input)

	if *input.RepositoryName == "repo1" {
		c.assertImageDigests(&[]string{"digest-tag1", "digest-tag2"}, &input.ImageIds)
		return new(ecr.BatchDeleteImageOutput), nil
	}

	if *input.RepositoryName == "repo2" {
		c.assertImageDigests(&[]string{"digest-tag3"}, &input.ImageIds)
		return new(ecr.BatchDeleteImageOutput), nil
	}

	return nil, fmt.Errorf("Unexpected repository: %s", *input.RepositoryName)
}

func (c *MockEcrClient) DeleteRepository(ctx context.Context, input *ecr.DeleteRepositoryInput, optFns ...func(*ecr.Options)) (response *ecr.DeleteRepositoryOutput, err error) {
	c.deleteRepoRequests = append(c.deleteRepoRequests, *input)

	if *input.RepositoryName == "repo1" || *input.RepositoryName == "repo3" {
		return &ecr.DeleteRepositoryOutput{}, nil
	}

	return nil, fmt.Errorf("Unexpected repository: %s", *input.RepositoryName)
}

func (c *MockEcrClient) DeleteLifecyclePolicy(ctx context.Context, input *ecr.DeleteLifecyclePolicyInput, optFns ...func(*ecr.Options)) (response *ecr.DeleteLifecyclePolicyOutput, err error) {
	c.deleteLifecycleRequests = append(c.deleteLifecycleRequests, *input)

	response = &ecr.DeleteLifecyclePolicyOutput{
		RegistryId:     aws.String(registryId),
		RepositoryName: input.RepositoryName,
	}
	return response, nil
}

func (e *MockEcrClient) assertCounts(t *testing.T, expected map[string]int) {
	countRequests := map[string]int{
		"describeRepos":         len(e.describeRepoRequests),
		"describeImages":        len(e.describeImageRequests),
		"batchDeleteImageRepos": len(e.batchDeleteImageRepoRequests),
		"deleteRepos":           len(e.deleteRepoRequests),
		"deleteLifecycles":      len(e.deleteLifecycleRequests),
	}
	for k, v := range countRequests {
		e := 0
		if val, ok := expected[k]; ok {
			e = val
			delete(expected, k)
		}
		if v != e {
			t.Errorf("Expected %d %s requests: %d", e, k, v)
		}
	}

	if len(expected) > 0 {
		t.Errorf("Invalid expected counts: %v", expected)
	}
}

func setup(t *testing.T) (*ecrDeletionHandler, *MockEcrClient) {
	ecrClient := MockEcrClient{
		t: t,
	}
	ecrH := ecrDeletionHandler{
		registryId:           registryId,
		expiresAfterPullDays: 2,
		client:               &ecrClient,
	}

	return &ecrH, &ecrClient
}

// Tests

func TestScanAndDeleteRepos(t *testing.T) {
	ecrH, ecrClient := setup(t)
	err := ecrH.ScanAndDeleteRepos(olderThanTime())
	if len(err) != 0 {
		t.Errorf("Unexpected error: %v", err)
	}

	ecrClient.assertCounts(t, map[string]int{
		"describeRepos":         1,
		"describeImages":        3,
		"batchDeleteImageRepos": 2,
		"deleteRepos":           2,
		"deleteLifecycles":      2,
	})

	// Check that the correct images and repos were deleted
	// - repo1: both images should be deleted, repo should be deleted
	// - repo2: first image should be deleted, second image should be kept
	// - repo3: no images, repo should be deleted

	r1 := ecrClient.batchDeleteImageRepoRequests[0]
	if *r1.RepositoryName != "repo1" {
		t.Errorf("batchDeleteImageRepoRequests[0]: Expected repo1, got %s", *r1.RepositoryName)
	}
	r1iids := r1.ImageIds
	if len(r1iids) != 2 {
		t.Errorf("batchDeleteImageRepoRequests[0]: Expected 2 image ids, got %d", len(r1iids))
	}
	ecrClient.assertImageDigests(&[]string{"digest-tag1", "digest-tag2"}, &r1iids)

	r2 := ecrClient.batchDeleteImageRepoRequests[1]
	if *r2.RepositoryName != "repo2" {
		t.Errorf("batchDeleteImageRepoRequests[0]: Expected repo2, got %s", *r2.RepositoryName)
	}
	r2iids := r2.ImageIds
	if len(r2iids) != 1 {
		t.Errorf("batchDeleteImageRepoRequests[1]: Expected 1 image ids, got %d", len(r1iids))
	}
	ecrClient.assertImageDigests(&[]string{"digest-tag3"}, &r2iids)

	d1 := ecrClient.deleteRepoRequests[0]
	if *d1.RepositoryName != "repo1" {
		t.Errorf("deleteRepoRequests[0]: Expected repo1, got %s", *d1.RepositoryName)
	}

	d2 := ecrClient.deleteRepoRequests[1]
	if *d2.RepositoryName != "repo3" {
		t.Errorf("deleteRepoRequests[0]: Expected repo3, got %s", *d2.RepositoryName)
	}
}

func TestDryRun(t *testing.T) {
	ecrH, ecrClient := setup(t)
	ecrH.dryRun = true
	err := ecrH.ScanAndDeleteRepos(olderThanTime())
	if len(err) != 0 {
		t.Errorf("Unexpected error: %v", err)
	}

	ecrClient.assertCounts(t, map[string]int{
		"describeRepos":         1,
		"describeImages":        3,
		"batchDeleteImageRepos": 0,
		"deleteRepos":           0,
		"deleteLifecycles":      0,
	})
}

func TestScanAndDeleteImages(t *testing.T) {
	ecrH, ecrClient := setup(t)
	err := ecrH.ScanAndDeleteImages("repo2", olderThanTime())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	ecrClient.assertCounts(t, map[string]int{
		"describeRepos":         0,
		"describeImages":        1,
		"batchDeleteImageRepos": 1,
		"deleteRepos":           0,
		"deleteLifecycles":      0,
	})
}

func TestDeleteImage(t *testing.T) {
	ecrH, ecrClient := setup(t)

	err := ecrH.DeleteImage("repo1", &[]string{"digest-tag1", "digest-tag2"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	ecrClient.assertCounts(t, map[string]int{
		"batchDeleteImageRepos": 1,
	})
}

func TestDeleteRepository(t *testing.T) {
	ecrH, ecrClient := setup(t)

	err := ecrH.DeleteRepository("repo3")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	ecrClient.assertCounts(t, map[string]int{
		"deleteRepos":      1,
		"deleteLifecycles": 1,
	})

	if *ecrClient.deleteRepoRequests[0].RepositoryName != "repo3" {
		t.Errorf("Unexpected deleteRepoRequests: %v", *ecrClient.deleteRepoRequests[0].RepositoryName)
	}
}
