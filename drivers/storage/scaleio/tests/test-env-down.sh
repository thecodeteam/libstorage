#!/usr/bin/env bash

# This script cleans up the infrastructure used for tests

set -e
# set -x

CF_STACK_NAME="$1"

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
  echo "Usage: ${0} stack-name"
  echo ""
  echo "   stack-name: AWS stack name. Must be uniquely identifiable"
}

if [ -z "${CF_STACK_NAME}" ]; then
  usage
  exit 1
fi

# Get stack ID
CF_STACK_ID=$(aws cloudformation describe-stacks \
  --stack-name ${CF_STACK_NAME} \
  --output text \
  --query 'Stacks[0].StackId')

# Delete cloud formation stack
aws cloudformation delete-stack \
  --stack-name ${CF_STACK_NAME}

echo "Waiting for CF stack to get deleted ..."

aws cloudformation wait stack-delete-complete \
  --stack-name ${CF_STACK_ID}

rm ~/.aws/config
rm ~/.aws/credentials

echo "Stack has been deleted"
