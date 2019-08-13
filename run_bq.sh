#!/bin/bash
yals bigquery \
-f=formida-parser \
-p=project_name \
-d=logs \
-t=test \
-s=secrets.json \
-l=/path/to/logfile