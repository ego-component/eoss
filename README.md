# AWOS: Wrapper For Aliyun OSS And Amazon S3

awos for node: https://github.com/shimohq/awos-js

## Features

- enable shards bucket
- add retry strategy
- avoid 404 status code:
  - `Get(objectName string) (string, error)` will return `"", nil` when object not exist
  - `Head(key string, meta []string) (map[string]string, error)` will return `nil, nil` when object not exist

## Installing

Use go get to retrieve the SDK to add it to your GOPATH workspace, or project's Go module dependencies.

```bash
go get github.com/ego-component/awos
```

## How to use
### config
```toml
[storage]
storageType = "oss" # oss|s3
accessKeyID = "xxx"
accessKeySecret = "xxx"
endpoint = "oss-cn-beijing.aliyuncs.com"
bucket = "aaa"
shards = []
[storage.buckets.template] # 可配置多套buckets配置，buckets里的配置会替换上层配置
bucket = "template-bucket"
shards = []
[storage.buckets.fileContent]
bucket = "contents-bucket"
shards = [
 "abcdefghijklmnopqr",
 "stuvwxyz0123456789"
]
```

```golang
import "github.com/ego-component/awos"

// 单独一个 bucket 配置
client := awos.Load("storage").Build()
// 多 bucket 配置
client := awos.Load("storage").Build(awos.WithBucketKey("template"))

// 带context（可记录链路）
client.WithContext(ctx).Get(key)
```

Available operations：

```golang
WithContext(ctx context.Context) Component
Get(key string, options ...GetOptions) (string, error)
GetBytes(key string, options ...GetOptions) ([]byte, error)
GetAsReader(key string, options ...GetOptions) (io.ReadCloser, error)
GetWithMeta(key string, attributes []string, options ...GetOptions) (io.ReadCloser, map[string]string, error)
Put(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error
Del(key string) error
DelMulti(keys []string) error
Head(key string, meta []string) (map[string]string, error)
ListObject(key string, prefix string, marker string, maxKeys int, delimiter string) ([]string, error)
SignURL(key string, expired int64) (string, error)
GetAndDecompress(key string) (string, error)
GetAndDecompressAsReader(key string) (io.ReadCloser, error)
CompressAndPut(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error
Range(key string, offset int64, length int64) (io.ReadCloser, error)
Exists(key string)(bool, error)
```
