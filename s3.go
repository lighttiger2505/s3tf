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

type S3ObjectType int

const (
	Bucket S3ObjectType = iota //0
	Dir
	Object
)

type S3Object struct {
	ObjType S3ObjectType
	Name    string
	Date    *time.Time
	Size    *int64
}

func NewS3Object(objType S3ObjectType, name string, date *time.Time, size *int64) *S3Object {
	return &S3Object{
		ObjType: objType,
		Name:    name,
		Date:    date,
		Size:    size,
	}
}

func ListBuckets() []*S3Object {
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

	var objects []*S3Object
	for _, bucket := range result.Buckets {
		obj := NewS3Object(
			Bucket,
			aws.StringValue(bucket.Name),
			bucket.CreationDate,
			nil,
		)
		objects = append(objects, obj)
	}
	return objects
}

func ListObjects(bucket, prefix string) []*S3Object {
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
		Prefix:    aws.String(prefix),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
		}
		os.Exit(1)
	}

	var objects []*S3Object
	for _, commonPrefix := range result.CommonPrefixes {
		obj := NewS3Object(
			Dir,
			aws.StringValue(commonPrefix.Prefix),
			nil,
			nil,
		)
		objects = append(objects, obj)
	}
	for _, content := range result.Contents {
		obj := NewS3Object(
			Object,
			aws.StringValue(content.Key),
			content.LastModified,
			content.Size,
		)
		objects = append(objects, obj)
	}
	return objects
}
