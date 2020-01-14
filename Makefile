export SPANNER_PROJECT_ID=yo-test-20202
export SPANNER_INSTANCE_ID=spanner-backup-test

deploy:
	GO111MODULE=on go mod vendor
	gcloud functions deploy SpannerBackupRunner --trigger-topic=spanner-backup --runtime=go111 --project $PROJECT_ID --timeout 300 --env-vars-file=./.env-mainnet-prod

run-backup:
	GO111MODULE=on go run main.go

integration-test-setup-db:
	./integrationtest/integration-test-setup.sh

integration-test-setup-data:
	GO111MODULE=on go run ./integrationtest/setup.go

yo:
	yo $(SPANNER_PROJECT_ID) $(SPANNER_INSTANCE_ID) database_1 -o integrationtest/yo
