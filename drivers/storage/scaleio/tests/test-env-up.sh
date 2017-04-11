#!/usr/bin/env bash

# This script launches infrastructure required for testing the ScaleIO storage driver.
# It spins up VPC and EC2 instance so AWS account owner will get charged for time
# that resources will be running.

#set -e

AWS_ACCESSKEY=${AWS_ACCESSKEY:-$1}
AWS_SECRETKEY=${AWS_SECRETKEY:-$2}
LAUNCH_KEY_NAME=${LAUNCH_KEY_NAME:-$3}
CF_STACK_NAME=${CF_STACK_NAME:-$4}

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

usage() {
  echo "Usage: ${0} access-key secret-key stack-name launch-key-name"
  echo ""
  echo "   access-key: AWS access key"
  echo "   secret-key: AWS secret key"
  echo "   launch-key-name: AWS key that will be used to launch EC2 instance."
  echo "                    The ssh key must be of the same filename in ~/.ssh"
  echo "   stack-name: AWS stack name. Must be uniquely identifiable"
}

template_path() {
  echo "$(dirname $0)/test-cf-template.json"
}

if [ -z "${AWS_ACCESSKEY}" ]; then
  usage
  exit 1
fi
if [ -z "${AWS_SECRETKEY}" ]; then
  usage
  exit 1
fi
if [ -z "${LAUNCH_KEY_NAME}" ]; then
  usage
  exit 1
fi
if [ -z "${CF_STACK_NAME}" ]; then
  CF_STACK_NAME=$(cat /dev/urandom | LC_CTYPE=C tr -dc '[:alpha:]' | head -c 16)
fi

if [ ! -f ~/.ssh/${LAUNCH_KEY_NAME}.pem ]; then
  curl -O https://dl.bintray.com/dvonthenen/libfake/dependencies/${LAUNCH_KEY_NAME}.pem
  cp -f ${LAUNCH_KEY_NAME}.pem ~/.ssh/
  chmod 600 ~/.ssh/${LAUNCH_KEY_NAME}.pem
fi

rm -f ./keyfile
if [ -f ~/.ssh/${LAUNCH_KEY_NAME}.pem ]; then
  echo ${LAUNCH_KEY_NAME}.pem > ./keyfile
else
  usage
  exit 1
fi

rm -f ./scaleio-uniquename
echo ${CF_STACK_NAME} > ./scaleio-uniquename

if [ ! -f ~/.aws/credentials ]; then
  aws configure set aws_access_key_id ${AWS_ACCESSKEY}
  aws configure set aws_secret_access_key ${AWS_SECRETKEY}
fi
if [ ! -f ~/.aws/config ]; then
  aws configure set default.region us-west-2
fi

# Launch CF stack
aws cloudformation create-stack \
  --stack-name ${CF_STACK_NAME} \
  --template-body file://$(template_path) \
  --parameter ParameterKey=KeyName,ParameterValue=${LAUNCH_KEY_NAME} \
  --capabilities CAPABILITY_IAM 1>/dev/null

echo "Environment launch started. It will take couple minutes to create whole environment..."
