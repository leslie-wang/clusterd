# complete
curl -s -X POST 'http://localhost:8088/mediaproc/v1/record?Action=CreateLiveRecordRule&DomainName=5000.livepush.myqcloud.com&AppName=live&StreamName=stream1&TemplateId=1000'

# no template id
curl -s -X POST 'http://localhost:8088/mediaproc/v1/record?Action=CreateLiveRecordRule&DomainName=5000.livepush.myqcloud.com&AppName=live&StreamName=stream1'

# no app name, have stream name
curl -s -X POST 'http://localhost:8088/mediaproc/v1/record?Action=CreateLiveRecordRule&DomainName=5000.livepush.myqcloud.com&StreamName=stream1'

# no stream name, have app name
curl -s -X POST 'http://localhost:8088/mediaproc/v1/record?Action=CreateLiveRecordRule&DomainName=5000.livepush.myqcloud.com&AppName=live'

# no stream name and app name
curl -s -X POST 'http://localhost:8088/mediaproc/v1/record?Action=CreateLiveRecordRule&DomainName=5000.livepush.myqcloud.com'
