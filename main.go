package spanner_backup_runner

import (
	"context"
	"golang.org/x/oauth2"
	dataflow "google.golang.org/api/dataflow/v1b3"
	"log"
	"os"
	"strings"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func SpannerBackupTrigger(ctx context.Context, m PubSubMessage) error {
	_ = string(m.Data)

	var (
		projectID = os.Getenv("PROJECT_ID")
		jobName = os.Getenv("JOB_NAME")
		jobLocation = os.Getenv("JOB_LOCATION")
		spannerInstanceID = os.Getenv("INSTANCE_ID")
		spannerDatabaseIDS = os.Getenv("DATABASE_IDS") // comma-separated databaseIDs
		backupOutputDir = os.Getenv("OUTPUT_DIR")
	)

	if projectID == "" ||
		jobName == "" ||
		jobLocation == "" ||
		spannerInstanceID == "" ||
		spannerDatabaseIDS == "" ||
		backupOutputDir == "" {
		log.Fatalf("param is not set. '%s', '%s', '%s', '%s', '%s', '%s'", projectID, jobName, jobLocation, spannerInstanceID, spannerDatabaseIDS, backupOutputDir)
	}


	// API doc: https://cloud.google.com/dataflow/docs/reference/rest/
	service, err := dataflow.NewService(oauth2.NoContext)
	if err != nil {
		log.Fatalf("err: %+v", err)
	}

	for _, spannerDatabaseID := range strings.Split(spannerDatabaseIDS, ",") {
		// Template docs:
		// https://cloud.google.com/dataflow/docs/guides/templates/provided-batch#cloudspannertogcsavro
		// https://github.com/GoogleCloudPlatform/DataflowTemplates/tree/master/src/main/java/com/google/cloud/teleport/templates
		job, err := dataflow.NewProjectsLocationsTemplatesService(service).Create(projectID, jobLocation, &dataflow.CreateJobFromTemplateRequest{
			GcsPath: "gs://dataflow-templates/latest/Cloud_Spanner_to_GCS_Avro",
			JobName: jobName + "-" + spannerDatabaseID,
			Parameters: map[string]string{ // parameter names are camel-case
				"instanceId": spannerInstanceID,
				"databaseId": spannerDatabaseID,
				"outputDir":  backupOutputDir,
			},
		}).Do()

		if err != nil {
			log.Printf("backup job error: %+v", err)
			continue
		}
		log.Printf("job is successfully started. job: %+v", job)
	}

	return nil
}