#!/usr/bin/env bash

# This script runs tests

# set -e

CF_STACK_NAME=${CF_STACK_NAME:-$1}
CF_EC2_USER=${CF_EC2_USER:-$2}
GIT_LIBSTORAGE_COMMIT_ID=${GIT_LIBSTORAGE_COMMIT_ID:-$3}
GIT_REXRAY_COMMIT_ID=${GIT_REXRAY_COMMIT_ID:-$4}

# Make sure that jq is installed
hash curl 2>/dev/null || {
  if [ -e "/etc/redhat-release" -o \
         -e "/etc/redhat-version" ]; then
    sudo yum -y install curl
  elif [ -e "/etc/debian-release" -o \
         -e "/etc/debian-version" -o \
         -e "/etc/lsb-release" ]; then
    sudo apt-get install -y curl
  else
    sudo brew install curl
  fi
}
# Make sure that aws cli is installed
hash aws 2>/dev/null || {
  curl "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" -o "/tmp/awscli-bundle.zip"
  unzip /tmp/awscli-bundle.zip
  sudo /tmp/awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws
}
# Make sure that jq is installed
hash jq 2>/dev/null || {
  if [ -e "/etc/redhat-release" -o \
         -e "/etc/redhat-version" ]; then
    sudo yum -y install jq
  elif [ -e "/etc/debian-release" -o \
         -e "/etc/debian-version" -o \
         -e "/etc/lsb-release" ]; then
    sudo apt-get install -y jq
  else
    sudo brew install jq
  fi
}

usage() {
  echo "Usage: ${0} stack-name git-libstorage-commit git-libstorage-commit ec2-user"
  echo ""
  echo "   stack-name: AWS stack name. Must be uniquely identifiable"
  echo "   git-libstorage-commit: Git Commit ID. Default: master"
  echo "   git-rexray-commit: Git Commit ID. Default: master"
  echo "   ec2-user: User to log into instance. Default (RHEL7): ec2-user"
}

if [ -z "${CF_STACK_NAME}" ]; then
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

echo "Waiting for CF stack to come up ..."

aws cloudformation wait stack-create-complete \
  --stack-name ${CF_STACK_NAME}

# Get IP address of EC2 machine where tests can be executed
EFS_NODE1=$(aws cloudformation describe-stack-resources --stack-name ${CF_STACK_NAME} | jq '.StackResources[] | select(.LogicalResourceId=="EfsNode1")' | jq -r .PhysicalResourceId)
#echo ${EFS_NODE1}

EFS_NODE1_FQDN=$(aws ec2 describe-instances --instance-ids ${EFS_NODE1} | jq -r '.Reservations[0].Instances[0].PublicDnsName')
#echo ${EFS_NODE1_FQDN}

ssh-keyscan -H ${EFS_NODE1_FQDN} >> ~/.ssh/known_hosts

rm -f $(dirname $0)/output1.txt

# ssh key to use
SSH_KEY_FILE=$(cat ./keyfile)

# Copy REX-Ray build to EC2 instance
scp -i ~/.ssh/${SSH_KEY_FILE} ./config-standalone.yml $CF_EC2_USER@$EFS_NODE1_FQDN:/tmp/config.yml
scp -i ~/.ssh/${SSH_KEY_FILE} ./tests1.sh $CF_EC2_USER@$EFS_NODE1_FQDN:/tmp
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$EFS_NODE1_FQDN "sudo rm -rf /tmp/finished.txt"
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$EFS_NODE1_FQDN "sudo chmod +x /tmp/tests1.sh"
if [ -f $GIT_LIBSTORAGE_COMMIT_ID ]; then
  scp -i ~/.ssh/${SSH_KEY_FILE} $GIT_LIBSTORAGE_COMMIT_ID $CF_EC2_USER@$EFS_NODE1_FQDN:/tmp
  ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$EFS_NODE1_FQDN "sudo chmod +x /tmp/rexray"
fi

if [ -f $GIT_LIBSTORAGE_COMMIT_ID ]; then
  GIT_LIBSTORAGE_COMMIT_ID="/tmp/rexray"
fi

echo "Executing Tests!"
#reduced EFS testing to single node because of API throttling issue on AWS

# Run tests... note the use of "bash --login -c". this forces a load of the user's
# env variables while cause the make command to fail.
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$EFS_NODE1_FQDN "bash --login -c '/tmp/tests1.sh $CF_STACK_NAME $GIT_LIBSTORAGE_COMMIT_ID $GIT_REXRAY_COMMIT_ID' 2>&1"

# Copy test coverage results
scp -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$EFS_NODE1_FQDN:/tmp/output.txt $(dirname $0)/output1.txt

echo "Tests passed and coverge results are available in the output files"
echo "Node 1 Results"
cat $(dirname $0)/output1.txt
