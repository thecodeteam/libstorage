#!/usr/bin/env bash

# This script runs tests

# set -e

: ${COVERPROFILE_NAME:="scaleio.test.out"}

TEST_BINARY=${TEST_BINARY:-$1}
CF_EC2_USER=${CF_EC2_USER:-$2}

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
  echo "Usage: ${0} test-binary ec2-user"
  echo ""
  echo "   test-binary: Path to compiled and runnable golang binary"
  echo "   ec2-user: User to log into instance. Default (RHEL7): ec2-user"
}

if [ -z "${TEST_BINARY}" ]; then
  usage
  exit 1
fi
if [ -z "${CF_EC2_USER}" ]; then
  CF_EC2_USER="ec2-user"
fi

CF_STACK_NAME=$(cat ./scaleio-uniquename)
if [ -z "${CF_STACK_NAME}" ]; then
  echo "stack-name not found"
  exit 1
fi

echo "Waiting for CF stack to come up ..."

aws cloudformation wait stack-create-complete \
  --stack-name ${CF_STACK_NAME}

# Get IP address of EC2 machine where tests can be executed
SCALEIO_NODE1=$(aws cloudformation describe-stack-resources --stack-name ${CF_STACK_NAME} | jq '.StackResources[] | select(.LogicalResourceId=="ScaleIONode1")' | jq -r .PhysicalResourceId)
#SCALEIO_NODE2=$(aws cloudformation describe-stack-resources --stack-name ${CF_STACK_NAME} | jq '.StackResources[] | select(.LogicalResourceId=="ScaleIONode2")' | jq -r .PhysicalResourceId)
#SCALEIO_NODE3=$(aws cloudformation describe-stack-resources --stack-name ${CF_STACK_NAME} | jq '.StackResources[] | select(.LogicalResourceId=="ScaleIONode3")' | jq -r .PhysicalResourceId)
#echo ${SCALEIO_NODE1}
#echo ${SCALEIO_NODE2}
#echo ${SCALEIO_NODE3}

SCALEIO_NODE1_FQDN=$(aws ec2 describe-instances --instance-ids ${SCALEIO_NODE1} | jq -r '.Reservations[0].Instances[0].PublicDnsName')
#SCALEIO_NODE2_FQDN=$(aws ec2 describe-instances --instance-ids ${SCALEIO_NODE2} | jq -r '.Reservations[0].Instances[0].PublicDnsName')
#SCALEIO_NODE3_FQDN=$(aws ec2 describe-instances --instance-ids ${SCALEIO_NODE3} | jq -r '.Reservations[0].Instances[0].PublicDnsName')
#echo ${SCALEIO_NODE1_FQDN}
#echo ${SCALEIO_NODE2_FQDN}
#echo ${SCALEIO_NODE3_FQDN}

# wait for the OS to be running
READY1=$(aws ec2 describe-instance-status --instance-ids ${SCALEIO_NODE1} | jq -r '.InstanceStatuses[0].SystemStatus.Status')
while [ "$READY1" != "ok" ];
do
  sleep 10
  echo "Checking ScaleIO Node1..."
  READY1=$(aws ec2 describe-instance-status --instance-ids ${SCALEIO_NODE1} | jq -r '.InstanceStatuses[0].SystemStatus.Status')
done
#READY2=$(aws ec2 describe-instance-status --instance-ids ${SCALEIO_NODE2} | jq -r '.InstanceStatuses[0].SystemStatus.Status')
#while [ "$READY2" != "ok" ];
#do
#  sleep 10
#  echo "Checking ScaleIO Node2..."
#  READY2=$(aws ec2 describe-instance-status --instance-ids ${SCALEIO_NODE2} | jq -r '.InstanceStatuses[0].SystemStatus.Status')
#done
#READY3=$(aws ec2 describe-instance-status --instance-ids ${SCALEIO_NODE3} | jq -r '.InstanceStatuses[0].SystemStatus.Status')
#while [ "$READY3" != "ok" ];
#do
#  sleep 10
#  echo "Checking ScaleIO Node3..."
#  READY3=$(aws ec2 describe-instance-status --instance-ids ${SCALEIO_NODE3} | jq -r '.InstanceStatuses[0].SystemStatus.Status')
#done

sleep 10

ssh-keyscan -H ${SCALEIO_NODE1_FQDN} >> ~/.ssh/known_hosts
#ssh-keyscan -H ${SCALEIO_NODE2_FQDN} >> ~/.ssh/known_hosts
#ssh-keyscan -H ${SCALEIO_NODE3_FQDN} >> ~/.ssh/known_hosts

FIRST_VOLUME=$CF_STACK_NAME"_1a"
SECOND_VOLUME=$CF_STACK_NAME"_1b"

# ssh key to use
SSH_KEY_FILE=$(cat ./keyfile)

# Copy go test binary to EC2 instance
scp -i ~/.ssh/${SSH_KEY_FILE} $TEST_BINARY $CF_EC2_USER@$SCALEIO_NODE1_FQDN:/tmp/scaleio.test
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE1_FQDN "sudo chmod +x /tmp/scaleio.test"

echo "Executing Tests!"
# Run tests... note the use of "bash --login -c". this forces a load of the user's
# env variables while cause the make command to fail.
# Yes, node1 uses test2 and vice versa
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE1_FQDN "bash --login -c 'FIRST_VOLUME=$FIRST_VOLUME SECOND_VOLUME=$SECOND_VOLUME sudo -E /tmp/scaleio.test -test.coverprofile ${COVERPROFILE_NAME}'"

# Copy test coverage results
scp -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE1_FQDN:$COVERPROFILE_NAME $(dirname $0)

echo "Tests passed and coverge results are available at ${COVERPROFILE_NAME}"
