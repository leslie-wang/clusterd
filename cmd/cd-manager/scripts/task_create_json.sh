current_timestamp=$(date +%s)

echo "{\"TemplateId\":1,\"DomainName\": \"test.play.com\",\"AppName\": \"live\",\"StreamName\":\"livetest\",\"RecordStreams\": [{\"SourceURL\": \"http://localhost:8000/output.mp4\"}],\"StorePath\" :\"/tmp/record\",\"EndTime\": $((current_timestamp + 6000))}" > /tmp/record_task.json
#echo "{\"TemplateId\":1,\"DomainName\": \"test.play.com\",\"AppName\": \"live\",\"StreamName\":\"livetest\",\"RecordStreams\": [{\"SourceURL\": \"udp://@:1234\"}],\"StorePath\" :\"/tmp/record\",\"EndTime\": $((current_timestamp + 6000))}" > /tmp/record_task.json
#echo "{\"TemplateId\":1,\"DomainName\": \"test.play.com\",\"AppName\": \"live\",\"StreamName\":\"livetest\",\"RecordStreams\": [{\"SourceURL\": \"udp://localhost:1234\"}],\"StorePath\" :\"/tmp/record\",\"EndTime\": $((current_timestamp + 6000)), \"Mp4FileDuration\": 120}" > /tmp/record_task.json
#echo "{\"TemplateId\":1,\"DomainName\": \"test.play.com\",\"AppName\": \"live\",\"StreamName\":\"livetest\",\"RecordStreams\": [{\"SourceURL\": \"rtp://127.0.0.1:50003\"},{\"SourceURL\": \"rtp://127.0.0.1:50001\"}],\"StorePath\" :\"/tmp/record\",\"EndTime\": $((current_timestamp + 6000)), \"Mp4FileDuration\": 120}" > /tmp/record_task.json
curl -s -X POST -H 'content-type: application//json' --data-binary @/tmp/record_task.json "http://localhost:8088/mediaproc/v1/record?Action=CreateRecordTask"

#curl -s -X POST -H 'content-type: application//json' --data-binary @./record_task.json "http://localhost:8088/mediaproc/v1/record?Action=CreateRecordTask"
