#!/usr/bin/env bash

# This script launches infrastructure required for testing the ScaleIO storage driver.
# It spins up VPC and EC2 instance so AWS account owner will get charged for time
# that resources will be running.

set -e
# set -x

AWS_ACCESS_KEY="$1"
AWS_SECRET_KEY="$2"
CF_STACK_NAME="$3"
LAUNCH_KEY_NAME="$4"

# Make sure that aws cli is installed
hash aws 2>/dev/null || {
  echo >&2 "Missing AWS command line. Please install aws cli: https://aws.amazon.com/cli/"
  exit 1
}
# Make sure that jq is installed
hash jq 2>/dev/null || {
  echo >&2 "Missing jq command line. Please install the jq utility"
  exit 1
}

usage() {
  echo "Usage: ${0} access-key secret-key stack-name launch-key-name"
  echo ""
  echo "   access-key: AWS access key"
  echo "   secret-key: AWS secret key"
  echo "   stack-name: AWS stack name. Must be uniquely identifiable"
  echo "   launch-key-name: AWS key that will be used to launch EC2 instance"
}

template_path() {
  echo "$(dirname $0)/test-cf-template.json"
}

if [ -z "${CF_STACK_NAME}" ]; then
  usage
  exit 1
fi
if [ -z "${LAUNCH_KEY_NAME}" ]; then
  usage
  exit 1
fi

aws configure set aws_access_key_id ${AWS_ACCESS_KEY}
aws configure set aws_secret_access_key ${AWS_SECRET_KEY}
aws configure set default.region us-west-2

# Launch CF stack
aws cloudformation create-stack \
  --stack-name ${CF_STACK_NAME} \
  --template-body file://$(template_path) \
  --parameter ParameterKey=KeyName,ParameterValue=${LAUNCH_KEY_NAME} \
  --capabilities CAPABILITY_IAM 1>/dev/null

echo "Environment launch started. It will take couple minutes to create whole environment..."
