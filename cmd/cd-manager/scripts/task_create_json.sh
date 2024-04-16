current_timestamp=$(date +%s)

echo "{\"TemplateId\":1,\"DomainName\": \"test.play.com\",\"AppName\": \"live\",\"StreamName\":\"livetest\",\"SourceURL\": \"http://localhost:8000/test.mp4\",\"StorePath\" :\"/tmp/record1\",\"EndTime\": $((current_timestamp + 60))}" > /tmp/record_task.json

curl -s -X POST -H 'content-type: application//json' --data-binary @/tmp/record_task.json "http://localhost:8088/mediaproc/v1/record?Action=CreateRecordTask"
