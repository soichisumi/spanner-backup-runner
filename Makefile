PROJECT_ID=your-project

deploy:
	GO111MODULE=on go mod vendor
	gcloud functions deploy SpannerBackupRunner --trigger-topic=spanner-backup --runtime=go111 --project $PROJECT_ID --timeout 300 --env-vars-file=./.env-mainnet-prod

