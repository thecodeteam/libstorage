#!/usr/bin/env bash

# This script runs tests

AWS_ACCESSKEY=${AWS_ACCESSKEY:-$1}
AWS_SECRETKEY=${AWS_SECRETKEY:-$2}
CF_STACK_NAME=${CF_STACK_NAME:-$3}
LAUNCH_KEY_NAME=${LAUNCH_KEY_NAME:-$4}
GIT_LIBSTORAGE_COMMIT_ID=${GIT_LIBSTORAGE_COMMIT_ID:-$5}
GIT_REXRAY_COMMIT_ID=${GIT_REXRAY_COMMIT_ID:-$6}
CF_EC2_USER=${CF_EC2_USER:-$7}

usage() {
  echo "Usage: ${0} access-key secret-key stack-name launch-key-name git-libstorage-commit git-rexray-commit ec2-user"
  echo ""
  echo "   access-key: AWS access key"
  echo "   secret-key: AWS secret key"
  echo "   stack-name: AWS stack name. Must be uniquely identifiable"
  echo "   launch-key-name: AWS key that will be used to launch EC2 instance"
  echo "   git-libstorage-commit: Git Commit ID. Default: master"
  echo "   git-rexray-commit: Git Commit ID. Default: master"
  echo "   ec2-user: User to log into instance. Default (RHEL7): ec2-user"
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
if [ -z "${GIT_LIBSTORAGE_COMMIT_ID}" ]; then
  GIT_LIBSTORAGE_COMMIT_ID="master"
fi
if [ -z "${GIT_REXRAY_COMMIT_ID}" ]; then
  GIT_REXRAY_COMMIT_ID="master"
fi
if [ -z "${CF_EC2_USER}" ]; then
  CF_EC2_USER="ec2-user"
fi

./test-env-up.sh $AWS_ACCESSKEY $AWS_SECRETKEY $CF_STACK_NAME $LAUNCH_KEY_NAME
sleep 30
./test-run-aws.sh $CF_STACK_NAME $CF_EC2_USER $GIT_LIBSTORAGE_COMMIT_ID $GIT_REXRAY_COMMIT_ID
./test-env-down.sh $CF_STACK_NAME
