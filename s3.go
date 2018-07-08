package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func ListBuckets() []string {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	ctx := context.Background()
	timeout := time.Second * 30
	var cancelFn func()
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}
	defer cancelFn()

	result, err := svc.ListBucketsWithContext(ctx, &s3.ListBucketsInput{})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
		}
		os.Exit(1)
	}

	buckets := []string{}
	for _, output := range result.Buckets {
		buckets = append(buckets, aws.StringValue(output.Name))
	}
	return buckets
}

func ListObjects(bucket string) []string {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	ctx := context.Background()
	timeout := time.Second * 30
	var cancelFn func()
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}
	defer cancelFn()

	result, err := svc.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
		Prefix:    aws.String("logs/"),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
		}
		os.Exit(1)
	}

	keys := []string{}
	for _, commonPrefix := range result.CommonPrefixes {
		keys = append(keys, aws.StringValue(commonPrefix.Prefix))
	}
	for _, content := range result.Contents {
		keys = append(keys, aws.StringValue(content.Key))
	}
	return keys
}
