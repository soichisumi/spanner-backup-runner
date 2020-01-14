#! /bin/bash

for v in `seq 1 10`
do
  export SPANNER_DATABASE_ID="database_${v}"
  echo "creating database. ID: ${SPANNER_DATABASE_ID}"
  wrench create --directory ./integrationtest
done