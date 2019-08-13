#!/bin/bash
yals influx \
-d=test \
-s=test \
-c=user:pwd@http://localhost:8086/test \
-l=/path/to/logfile \
-f=formida-parser