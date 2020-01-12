package spannerutil

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/api/spanner/v1"
)

func ListAllDatabases(service *spanner.Service, projectID, instanceID string) ([]*spanner.Database, error) {
	databases := make([]*spanner.Database, 0, 0)
	nextPageToken := ""
	for {
		// doc: https://cloud.google.com/spanner/docs/reference/rest/?hl=ja#rest-resource-v1projectsinstancesdatabases
		res, err := spanner.NewProjectsInstancesDatabasesService(service).List(fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID)).PageSize(20).PageToken(nextPageToken).Do()
		if err != nil {
			fmt.Errorf("backupjob.listdatabases error: %+v\n", err)
			return nil, err
		}
		databases = append(databases, res.Databases...)
		if res.NextPageToken == "" {
			break
		}
		nextPageToken = res.NextPageToken
	}
	return databases, nil
}

func ListTargetDatabases(service *spanner.Service, ignoreDatabasesRegex *regexp.Regexp) ([]string, error) {
	databases, err := ListAllDatabases(service)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(databases))
	for _, v := range databases {
		dbName := strings.Split(v.Name, "/")[6] // form: `projects/<project>/instances/<instance>/databases/<database>`
		if ignoreDatabasesRegex.MatchString(dbName) {
			res = append(res, dbName)
		}
	}
	return res, nil
}
