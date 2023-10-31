curl -s -X POST "http://localhost:8088/mediaproc/v1/record?Action=DeleteLiveRecordTemplate&TemplateId=$1" | python3 -m json.tool

#!/bin/bash

# Make a GET request using curl and store the response in a variable
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:8088/mediaproc/v1/record?Action=DeleteLiveRecordTemplate&TemplateId=$1")

# Check if the HTTP status code is 200 (OK)
if [ "$response" -eq 200 ]; then
    # If status code is 200, fetch and parse the JSON response
    curl -s -X POST "http://localhost:8088/mediaproc/v1/record?Action=DeleteLiveRecordTemplate&TemplateId=$1" | python3 -m json.tool
else
    curl -s -X POST "http://localhost:8088/mediaproc/v1/record?Action=DeleteLiveRecordTemplate&TemplateId=$1"
fi
