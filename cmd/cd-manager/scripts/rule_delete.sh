#!/bin/bash

curl -s -X POST "http://localhost:8088/mediaproc/v1/record?Action=DeleteLiveRecordRule&DomainName=$1&AppName=$2&StreamName=$3"
