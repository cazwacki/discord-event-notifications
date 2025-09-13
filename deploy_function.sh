#!/bin/bash
set -e

go mod tidy
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap -tags lambda.norpc notify.go
zip lambda-handler.zip bootstrap
rm bootstrap
aws lambda update-function-code --function-name discord-event-notifications --zip-file fileb://lambda-handler.zip --region us-east-1
rm lambda-handler.zip