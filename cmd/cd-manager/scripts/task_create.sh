# complete
#curl -s -X POST 'http://localhost:8088/mediaproc/v1/record?Action=CreateRecordTask&AppName=live&DomainName=5000.live.push.com&StreamName=livetest&StartTime=1589889600&EndTime=1589904000&TemplateId=0'
# aHR0cDovL2xvY2FsaG9zdDo4MDAwL3Rlc3QubXA0 = http://localhost:8000/test.mp4
# start now
#curl -s -X POST 'http://localhost:8088/mediaproc/v1/record?Action=CreateRecordTask&DomainName=aHR0cDovL2xvY2FsaG9zdDo4MDAwL3Rlc3QubXA0&EndTime=1589904000&TemplateId=0'
# start in 1 minute
startTime=`echo $(($(date +%s) + 60))`
curl -s -X POST "http://localhost:8088/mediaproc/v1/record?Action=CreateRecordTask&DomainName=aHR0cDovL2xvY2FsaG9zdDo4MDAwL3Rlc3QubXA0&StartTime=$startTime&EndTime=1589904000&TemplateId=0"