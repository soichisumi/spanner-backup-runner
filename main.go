package main

import (
	"context"
	_ "encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/soichisumi/go-util/slice"

	"github.com/soichisumi/spanner-backup-runner/pkg/storageutil"

	"cloud.google.com/go/storage"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/soichisumi/go-util/logger"
	"github.com/soichisumi/spanner-backup-runner/pkg/dataflowutil"
	"github.com/soichisumi/spanner-backup-runner/pkg/spannerutil"
	"go.uber.org/zap"
	dataflow "google.golang.org/api/dataflow/v1b3"
	"google.golang.org/api/spanner/v1"
)

var (
	cfg                 Config
	dataflowService     *dataflow.Service
	spannerService      *spanner.Service
	storageClient       *storage.Client
	ignoreDatabaseRegex *regexp.Regexp
	ctx                 = context.Background()
)

type Config struct {
	ProjectID           string `envconfig:"PROJECT_ID" required:"true"`
	JobNamePrefix       string `envconfig:"JOB_NAME_PREFIX" required:"true"`
	JobLocation         string `envconfig:"JOB_LOCATION" required:"true"`
	InstanceID          string `envconfig:"INSTANCE_ID" required:"true"`
	OutputBucketName    string `envconfig:"OUTPUT_BUCKET_NAME" required:"true"`
	IgnoreDatabaseRegex string `envconfig:"IGNORE_DATABASE_REGEX" default:"^0"` // matches nothing
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file. err: %+v", err)
	}
	envconfig.MustProcess("", &cfg)

	s, err := dataflow.NewService(ctx)
	if err != nil {
		log.Fatalf("err: %+v", err)
	}
	dataflowService = s

	ss, err := spanner.NewService(ctx)
	if err != nil {
		log.Fatalf("err: %+v", err)
	}
	spannerService = ss

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("err: %+v", err)
	}
	storageClient = client

	ignoreDatabaseRegex = regexp.MustCompile(cfg.IgnoreDatabaseRegex)
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

	databaseIDs, err := spannerutil.ListDatabaseIDsWithFilter(spannerService, cfg.ProjectID, cfg.InstanceID, ignoreDatabaseRegex)
	if err != nil {
		logger.Error(err.Error(), zap.Error(err))
		return
	}
	backedUpDatabases, err := storageutil.ListObjectsInDirectory(storageClient, cfg.OutputBucketName, time.Now().Format("2006-01-02"))
	if err != nil {
		logger.Error(err.Error(), zap.Error(err))
		return
	}
	targetDatabaseIDs := make([]string, 0, 0)
	for _, v := range databaseIDs {
		if !slice.Contains(backedUpDatabases, v) {
			logger.Info("database is already backed up. skip.", zap.String("databaseID", v))
			continue
		}
		targetDatabaseIDs = append(targetDatabaseIDs, v)
	}

	logger.Info("target databaseIDs of backup", zap.Any("targetDatabaseIDs", targetDatabaseIDs))

	// api doc: https://cloud.google.com/dataflow/docs/reference/rest/
	// template doc: https://cloud.google.com/dataflow/docs/guides/templates/provided-batch#cloudspannertogcsavro
	for _, databaseID := range targetDatabaseIDs {
		job, err := dataflowutil.ExecuteSpannerBackupJob(dataflowService, cfg.ProjectID, cfg.JobLocation, cfg.JobNamePrefix, cfg.InstanceID, databaseID, cfg.OutputBucketName)
		if err != nil {
			logger.Error(err.Error(), zap.Error(err))
			continue
		}
		logger.Info("job is successfully started", zap.Any("job", job))
	}
	logger.Info("backup is successfully started")
	return
}

func main() {
	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Info(fmt.Sprintf("backup runner is listening on port %s ...", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
