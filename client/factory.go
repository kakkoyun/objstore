// Copyright (c) The Thanos Authors.
// Licensed under the Apache License 2.0.

package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"github.com/thanos-io/objstore"
	"github.com/thanos-io/objstore/azure"
	"github.com/thanos-io/objstore/bos"
	"github.com/thanos-io/objstore/cos"
	"github.com/thanos-io/objstore/filesystem"
	"github.com/thanos-io/objstore/gcs"
	"github.com/thanos-io/objstore/oss"
	"github.com/thanos-io/objstore/s3"
	"github.com/thanos-io/objstore/swift"
)

type ObjProvider string

const (
	FILESYSTEM ObjProvider = "FILESYSTEM"
	GCS        ObjProvider = "GCS"
	S3         ObjProvider = "S3"
	AZURE      ObjProvider = "AZURE"
	SWIFT      ObjProvider = "SWIFT"
	COS        ObjProvider = "COS"
	ALIYUNOSS  ObjProvider = "ALIYUNOSS"
	BOS        ObjProvider = "BOS"
)

type BucketConfig struct {
	Type   ObjProvider `yaml:"type"`
	Config interface{} `yaml:"config"`
}

// NewBucket initializes and returns new object storage clients.
// NOTE: confContentYaml can contain secrets.
func NewBucket(logger log.Logger, confContentYaml []byte, component string) (objstore.Bucket, error) {
	level.Info(logger).Log("msg", "loading bucket configuration")
	bucketConf := &BucketConfig{}
	if err := yaml.UnmarshalStrict(confContentYaml, bucketConf); err != nil {
		return nil, errors.Wrap(err, "parsing config YAML file")
	}

	config, err := yaml.Marshal(bucketConf.Config)
	if err != nil {
		return nil, errors.Wrap(err, "marshal content of bucket configuration")
	}

	var bucket objstore.Bucket
	switch strings.ToUpper(string(bucketConf.Type)) {
	case string(GCS):
		bucket, err = gcs.NewBucket(context.Background(), logger, config, component)
	case string(S3):
		bucket, err = s3.NewBucket(logger, config, component)
	case string(AZURE):
		bucket, err = azure.NewBucket(logger, config, component)
	case string(SWIFT):
		bucket, err = swift.NewContainer(logger, config)
	case string(COS):
		bucket, err = cos.NewBucket(logger, config, component)
	case string(ALIYUNOSS):
		bucket, err = oss.NewBucket(logger, config, component)
	case string(FILESYSTEM):
		bucket, err = filesystem.NewBucketFromConfig(config)
	case string(BOS):
		bucket, err = bos.NewBucket(logger, config, component)
	default:
		return nil, errors.Errorf("bucket with type %s is not supported", bucketConf.Type)
	}
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("create %s client", bucketConf.Type))
	}
	return bucket, nil
}
