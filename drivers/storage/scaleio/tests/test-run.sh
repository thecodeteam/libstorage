#!/usr/bin/env bash

# This script runs tests

# set -e
# set -x

: ${CF_EC2_USER:="ec2-user"}

CF_STACK_NAME="$1"
TEST_BINARY="$2"

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
  echo "Usage: ${0} stack-name rexray-path"
  echo ""
  echo "   stack-name: AWS stack name. Must be uniquely identifiable"
  echo "   rexray-path: Path to compiled and runnable golang binary"
}

if [ -z "${CF_STACK_NAME}" ]; then
  usage
  exit 1
fi
if [ -z "${TEST_BINARY}" ]; then
  usage
  exit 1
fi

# Require valid test binary file
if [ ! -f "${TEST_BINARY}" ]; then
  echo >&2 "${TEST_BINARY} is not a valid file"
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

ssh-keyscan -H ${SCALEIO_NODE1_FQDN} >> ~/.ssh/known_hosts
#ssh-keyscan -H ${SCALEIO_NODE2_FQDN} >> ~/.ssh/known_hosts
#ssh-keyscan -H ${SCALEIO_NODE3_FQDN} >> ~/.ssh/known_hosts

# Copy REX-Ray build to EC2 instance
ssh $CF_EC2_USER@$SCALEIO_NODE1_FQDN "sudo mkdir -p /usr/bin"
ssh $CF_EC2_USER@$SCALEIO_NODE1_FQDN "sudo mkdir -p /etc/rexray"
scp ${TEST_BINARY} $CF_EC2_USER@$SCALEIO_NODE1_FQDN:/tmp
scp ./config.yml $CF_EC2_USER@$SCALEIO_NODE1_FQDN:/tmp
scp ./tests.sh $CF_EC2_USER@$SCALEIO_NODE1_FQDN:/tmp
ssh $CF_EC2_USER@$SCALEIO_NODE1_FQDN "sudo cp -f /tmp/rexray /usr/bin"
ssh $CF_EC2_USER@$SCALEIO_NODE1_FQDN "sudo cp -f /tmp/config.yml /etc/rexray"
ssh $CF_EC2_USER@$SCALEIO_NODE1_FQDN "sudo chmod +x /tmp/tests.sh"
ssh $CF_EC2_USER@$SCALEIO_NODE1_FQDN "sudo /usr/bin/rexray service restart"

#ssh $CF_EC2_USER@$SCALEIO_NODE2_FQDN "sudo mkdir -p /usr/bin"
#ssh $CF_EC2_USER@$SCALEIO_NODE2_FQDN "sudo mkdir -p /etc/rexray"
#scp ${TEST_BINARY} $CF_EC2_USER@$SCALEIO_NODE2_FQDN:/tmp
#scp ./config.yml $CF_EC2_USER@$SCALEIO_NODE2_FQDN:/tmp
#scp ./tests.sh $CF_EC2_USER@$SCALEIO_NODE2_FQDN:/tmp
#ssh $CF_EC2_USER@$SCALEIO_NODE2_FQDN "sudo /tmp/rexray /usr/bin"
#ssh $CF_EC2_USER@$SCALEIO_NODE2_FQDN "sudo /tmp/config.yml /etc/rexray"
#ssh $CF_EC2_USER@$SCALEIO_NODE2_FQDN "sudo chmod +x /tmp/tests.sh"
#ssh $CF_EC2_USER@$SCALEIO_NODE2_FQDN "sudo /usr/bin/rexray service restart"

#ssh $CF_EC2_USER@$SCALEIO_NODE3_FQDN "sudo mkdir -p /usr/bin"
#ssh $CF_EC2_USER@$SCALEIO_NODE3_FQDN "sudo mkdir -p /etc/rexray"
#scp ${TEST_BINARY} $CF_EC2_USER@$SCALEIO_NODE3_FQDN:/tmp
#scp ./config.yml $CF_EC2_USER@$SCALEIO_NODE3_FQDN:/tmp
#scp ./tests.sh $CF_EC2_USER@$SCALEIO_NODE3_FQDN:/tmp
#ssh $CF_EC2_USER@$SCALEIO_NODE3_FQDN "sudo /tmp/rexray /usr/bin"
#ssh $CF_EC2_USER@$SCALEIO_NODE3_FQDN "sudo /tmp/config.yml /etc/rexray"
#ssh $CF_EC2_USER@$SCALEIO_NODE3_FQDN "sudo chmod +x /tmp/tests.sh"
#ssh $CF_EC2_USER@$SCALEIO_NODE3_FQDN "sudo /usr/bin/rexray service restart"

# Run tests
ssh $CF_EC2_USER@$SCALEIO_NODE1_FQDN "sudo /tmp/tests.sh"

# Copy test coverage results
scp $CF_EC2_USER@$SCALEIO_NODE1_FQDN:/tmp/output.txt $(dirname $0)

echo "Tests passed and coverge results are available at $(dirname $0)/output.txt"
