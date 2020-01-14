package main

import (
	"context"
	"math/rand"
	"strconv"

	"cloud.google.com/go/spanner"

	"github.com/google/uuid"
	"github.com/soichisumi/spanner-backup-runner/integrationtest/yo"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/soichisumi/go-gcp-util/spannerutil"
	"github.com/soichisumi/go-util/logger"
	"go.uber.org/zap"
)

var (
	cfg Config
)

type Config struct {
	ProjectID  string `envconfig:"PROJECT_ID" required:"true"`
	InstanceID string `envconfig:"INSTANCE_ID" required:"true"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal(err.Error(), zap.Error(err))
	}

	envconfig.MustProcess("", &cfg)
}

func insertTestdata(client *spanner.Client) error {
	ctx := context.Background()
	ms := make([]*spanner.Mutation, 0, 100)

	for j := 0; j < 100; j++ {
		r := rand.Int()
		data := yo.Testdatum{
			ID:  uuid.New().String(),
			Str: strconv.Itoa(r),
			Num: int64(r),
		}
		ms = append(ms, data.Insert(ctx))
	}
	_, err := client.Apply(ctx, ms)
	return err
}

func main() {

	for dbNum := 1; dbNum <= 10; dbNum++ {
		client, err := spannerutil.NewClient(spannerutil.Config{
			Project:  cfg.ProjectID,
			Instance: cfg.InstanceID,
			Database: "database_" + strconv.Itoa(dbNum),
		})
		if err != nil {
			logger.Error(err.Error(), zap.Error(err))
			continue
		}
		err = insertTestdata(client)
		if err != nil {
			logger.Error(err.Error(), zap.Error(err))
		}
	}
}
