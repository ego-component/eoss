package awos

type BuildOption func(c *Container)

func WithStorageType(storageType string) BuildOption {
	return func(c *Container) {
		c.config.StorageType = storageType
	}
}

func WithAccessKeyID(ak string) BuildOption {
	return func(c *Container) {
		c.config.AccessKeyID = ak
	}
}

func WithAccessKeySecret(sk string) BuildOption {
	return func(c *Container) {
		c.config.AccessKeySecret = sk
	}
}

func WithEndpoint(endpoint string) BuildOption {
	return func(c *Container) {
		c.config.Endpoint = endpoint
	}
}

func WithBucket(bucket string) BuildOption {
	return func(c *Container) {
		c.config.Bucket = bucket
	}
}

func WithShards(shards []string) BuildOption {
	return func(c *Container) {
		c.config.Shards = shards
	}
}

func WithRegion(region string) BuildOption {
	return func(c *Container) {
		c.config.Region = region
	}
}

func WithS3ForcePathStyle(s3ForcePathStyle bool) BuildOption {
	return func(c *Container) {
		c.config.S3ForcePathStyle = s3ForcePathStyle
	}
}

func WithSSL(ssl bool) BuildOption {
	return func(c *Container) {
		c.config.SSL = ssl
	}
}

func WithS3HttpTimeoutSecs(s3HttpTimeoutSecs int64) BuildOption {
	return func(c *Container) {
		c.config.S3HttpTimeoutSecs = s3HttpTimeoutSecs
	}
}

func WithBucketKey(bucketKey string) BuildOption {
	return func(c *Container) {
		c.config.bucketKey = bucketKey
	}
}
