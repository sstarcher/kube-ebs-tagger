package tagger

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var sess *session.Session

func init() {
	var err error

	metaSession, _ := session.NewSession()
	metaClient := ec2metadata.New(metaSession)
	region, _ := metaClient.Region()

	sess, err = session.NewSession(&aws.Config{
		Region: &region,
	})
	if err != nil {
		log.Fatalf("%v", err)
	}

}

// Tag syncs ebs volume tags
func Tag(volumeID string, labels map[string]string) (bool, error) {
	svc := ec2.New(sess)
	input := &ec2.DescribeVolumesInput{
		VolumeIds: []*string{&volumeID},
	}

	result, err := svc.DescribeVolumes(input)
	if err != nil {
		return false, err
	}

	if len(result.Volumes) != 1 {
		return false, errors.New("expected query to only return a single volume")
	}

	noChange := true
	for key, val := range labels {
		isMissing := true
		for _, tag := range result.Volumes[0].Tags {
			if *tag.Key != key {
				continue
			}

			if val != *tag.Value {
				noChange = false
			}
			isMissing = false
			break
		}

		if isMissing || !noChange {
			noChange = false
			break
		}
	}

	if noChange {
		return false, nil
	}

	tagInput := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(volumeID),
		},
	}

	tags := []*ec2.Tag{}
	for key, val := range labels {
		tag := ec2.Tag{
			Key:   aws.String(key),
			Value: aws.String(val),
		}
		tags = append(tags, &tag)
	}

	tagInput.Tags = tags
	_, err = svc.CreateTags(tagInput)
	if err != nil {
		return false, err
	}
	return true, nil
}
