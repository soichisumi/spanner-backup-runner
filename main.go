package main

import (
	"context"
	_ "encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/soichisumi/go-util/logger"
	"github.com/soichisumi/spanner-backup-runner/pkg/dataflowutil"
	"github.com/soichisumi/spanner-backup-runner/pkg/spannerutil"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	dataflow "google.golang.org/api/dataflow/v1b3"
	"google.golang.org/api/spanner/v1"
)

var (
	cfg                 Config
	dataflowService     *dataflow.Service
	spannerService      *spanner.Service
	ignoreDatabaseRegex *regexp.Regexp
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file. err: %+v", err)
	}
	err = envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalf("Error loading envvals. err: %+v", err)
	}

	s, err := dataflow.NewService(oauth2.NoContext)
	if err != nil {
		log.Fatalf("err: %+v", err)
	}
	dataflowService = s

	ss, err := spanner.NewService(context.Background())
	if err != nil {
		log.Fatalf("err: %+v", err)
	}
	spannerService = ss

	ignoreDatabaseRegex = regexp.MustCompile(cfg.IgnoreDatabaseRegex)
	if cfg.IgnoreDatabaseRegex == "" {
		ignoreDatabaseRegex = regexp.MustCompile(`0^`) // match nothing
	}
}

type Config struct {
	ProjectID           string `envconfig:"PROJECT_ID" required:"true"`
	JobNamePrefix       string `envconfig:"JOB_NAME_PREFIX" required:"true"`
	JobLocation         string `envconfig:"JOB_LOCATION" required:"true"`
	InstanceID          string `envconfig:"INSTANCE_ID" required:"true"`
	OutputBucketName    string `envconfig:"OUTPUT_BUCKET_NAME" required:"true"`
	IgnoreDatabaseRegex string `envconfig:"IGNORE_DATABASE_REGEX"`
}

type PubSubMessage struct {
	Message struct {
		Data []byte `json:"data,omitempty"`
		ID   string `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	_ = r.Body // discard request body

	targetDatabaseIDs, err := spannerutil.ListTargetDatabases(spannerService, ignoreDatabaseRegex)
	if err != nil {
		logger.Error(err.Error(), zap.Error(err))
		return
	}

	// api doc: https://cloud.google.com/dataflow/docs/reference/rest/
	// template doc: https://cloud.google.com/dataflow/docs/guides/templates/provided-batch#cloudspannertogcsavro
	for _, databaseID := range targetDatabaseIDs {
		job, err := dataflowutil.ExecuteSpannerBackupJob(dataflowService, cfg.ProjectID, cfg.JobLocation, cfg.JobNamePrefix, cfg.InstanceID, databaseID, cfg.OutputBucketName)
		if err != nil {
			logger.Error(err.Error(), zap.Error(err))
			continue
		}
		logger.Info("job is successfully started", zap.Any("jon", job))
	}

	return
}

func main() {
	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
}
