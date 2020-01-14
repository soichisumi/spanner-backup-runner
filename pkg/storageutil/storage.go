package storageutil

import (
	"context"
	"strings"

	"github.com/soichisumi/go-util/logger"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/storage"
)

var (
	ctx = context.Background()
)

func ListObjectsInDirectory(client *storage.Client, bucketName, objectPrefix string) ([]string, error) {
	it := client.Bucket(bucketName).Objects(ctx, &storage.Query{
		Delimiter: "/",
		Prefix:    objectPrefix,
		Versions:  false,
	})
	res := make([]string, 0, 0)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Error(err.Error(), zap.Error(err))
			return nil, err
		}
		objName := strings.ReplaceAll(attrs.Prefix, objectPrefix, "")
		res = append(res, objName)
	}
	return res, nil
}
