#!/bin/bash

curl -s -X POST "http://localhost:8088/mediaproc/v1/record?Action=DeleteRecordTask&TaskId=$1"
