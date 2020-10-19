#!/usr/bin/env bash
FIXTURE_PATH=./data/traces
FILE_ID="15ub8SQtHwt3S2W1niHGr2_8lFtDw-JXi"
if [[ -f ${FIXTURE_PATH} ]]; then
   echo "File $FIXTURE_PATH exists, skipping download."
else
    mkdir -p ./data
    curl -L -o ./data/traces https://www.dropbox.com/s/dyuk2b9aeysuznq/traces?dl=1
fi