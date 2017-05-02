#!/usr/bin/env bash

# This script runs tests

AWS_ACCESSKEY=${AWS_ACCESSKEY:-$1}
AWS_SECRETKEY=${AWS_SECRETKEY:-$2}
CF_STACK_NAME=${CF_STACK_NAME:-$3}
LAUNCH_KEY_NAME=${LAUNCH_KEY_NAME:-$4}
GIT_COMMIT_ID=${GIT_COMMIT_ID:-$5}

usage() {
  echo "Usage: ${0} access-key secret-key stack-name launch-key-name"
  echo ""
  echo "   access-key: AWS access key"
  echo "   secret-key: AWS secret key"
  echo "   stack-name: AWS stack name. Must be uniquely identifiable"
  echo "   launch-key-name: AWS key that will be used to launch EC2 instance"
  echo "   git-commit: Git Commit ID. Default: master"
}

if [ -z "${AWS_ACCESSKEY}" ]; then
  usage
  exit 1
fi
if [ -z "${AWS_SECRETKEY}" ]; then
  usage
  exit 1
fi
if [ -z "${CF_STACK_NAME}" ]; then
  usage
  exit 1
fi
if [ -z "${LAUNCH_KEY_NAME}" ]; then
  usage
  exit 1
fi
if [ -z "${GIT_COMMIT_ID}" ]; then
  GIT_COMMIT_ID="master"
fi

./test-env-up.sh $AWS_ACCESSKEY $AWS_SECRETKEY $CF_STACK_NAME $LAUNCH_KEY_NAME
sleep 5
./test-run-aws.sh $CF_STACK_NAME $GIT_COMMIT_ID
./test-env-down.sh $CF_STACK_NAME
