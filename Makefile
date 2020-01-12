export SPANNER_PROJECT_ID=your-project-id
export SPANNER_INSTANCE_ID=your-instance-id

FROM := 1
TO := 10
NUMBERS := $(shell seq ${FROM} ${TO})

deploy:
	GO111MODULE=on go mod vendor
	gcloud functions deploy SpannerBackupRunner --trigger-topic=spanner-backup --runtime=go111 --project $PROJECT_ID --timeout 300 --env-vars-file=./.env-mainnet-prod

integration-test-setup:
	for v in ${NUMBERS}; do \
		vv := "database_${v}" \
		SPANNER_DATABASE_ID=${vv} wrench create --directory ./\
	done
	GO111MODULE=on go run setup.go

