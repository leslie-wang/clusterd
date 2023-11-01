# complete
./rule_delete.sh 5000.livepush.myqcloud.com live stream1

# no app name, have stream name
./rule_delete.sh 5000.livepush.myqcloud.com "" stream1

# no stream name, have app name
./rule_delete.sh 5000.livepush.myqcloud.com live ""

# no stream name and app name
./rule_delete.sh 5000.livepush.myqcloud.com "" ""
