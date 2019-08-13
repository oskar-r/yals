#!/bin/bash
yals bigquery \
-r='(?m)(?P<date>20\d{2}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2})\s(?P<method>[A-Z]{1,10})\s{0,}\|(?P<route>[A-Za-z0-9=?&/_]{0,})\s{0,}\|(?P<respose>[0-9]{3})\s{0,}\|(?P<resptime>[0-9\.]{0,}).*\s{0,}\|size:(?P<size>[0-9]{0,})\s{0,}B\s{0,}\[request_id:(?P<req_id>[0-9a-z-]{0,})\s{0,}user:(?P<subject>[0-9-]{0,})\s{0,}role:(?P<role>[a-z\s]{0,})]' \
-df='2006/01/02 15:04:05' \
-p=project_name \
-d=logs \
-t=test \
-s=secrets.json \
-l=/path/to/logfile