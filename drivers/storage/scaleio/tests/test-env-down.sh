#!/usr/bin/env bash

# This script cleans up the infrastructure used for tests

#set -e

# Make sure that jq is installed
hash curl 2>/dev/null || {
  if [ -e "/etc/redhat-release" -o \
         -e "/etc/redhat-version" ]; then
    yum -y install curl
  elif [ -e "/etc/debian-release" -o \
         -e "/etc/debian-version" -o \
         -e "/etc/lsb-release" ]; then
    apt-get install -y curl
  else
    brew install curl
  fi
}
# Make sure that aws cli is installed
hash aws 2>/dev/null || {
  curl "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" -o "/tmp/awscli-bundle.zip"
  unzip /tmp/awscli-bundle.zip
  ./awscli-bundle/install -b ~/bin/aws
  export PATH=~/bin:$PATH
}
# Make sure that jq is installed
hash jq 2>/dev/null || {
  if [ -e "/etc/redhat-release" -o \
         -e "/etc/redhat-version" ]; then
    yum -y install jq
  elif [ -e "/etc/debian-release" -o \
         -e "/etc/debian-version" -o \
         -e "/etc/lsb-release" ]; then
    apt-get install -y jq
  else
    brew install jq
  fi
}

CF_STACK_NAME=$(cat ./scaleio-uniquename)
if [ -z "${CF_STACK_NAME}" ]; then
  echo "stack-name not found or already deleted"
  exit 0
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

rm -f ./scaleio-uniquename

echo "Stack has been deleted"
