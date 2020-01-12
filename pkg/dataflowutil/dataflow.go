package dataflowutil

import (
	dataflow "google.golang.org/api/dataflow/v1b3"
	"time"
)

func ExecuteSpannerBackupJob(service *dataflow.Service, projectID, location, jobNamePrefix, instanceID, databaseID, outputBucketName string) (*dataflow.Job, error) {
	return dataflow.NewProjectsLocationsTemplatesService(service).Create(projectID, location, &dataflow.CreateJobFromTemplateRequest{
		GcsPath: "gs://dataflow-templates/latest/Cloud_Spanner_to_GCS_Avro", //gs://dataflow-templates/2019-07-10-00/
		JobName: jobNamePrefix + "-" + databaseID,
		Parameters: map[string]string{ // parameter names are camel-case
			"instanceId": instanceID,
			"databaseId": databaseID,
			"outputDir":  outputBucketName + "/" + time.Now().Format("2006-01-02") + "/" + databaseID,
		},
	}).Do()
}