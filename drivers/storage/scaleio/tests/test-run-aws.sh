#!/usr/bin/env bash

# This script runs tests

# set -e

: ${CF_EC2_USER:="ec2-user"}

CF_STACK_NAME=${CF_STACK_NAME:-$1}
GIT_COMMIT_ID=${GIT_COMMIT_ID:-$2}

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
  echo "Usage: ${0} stack-name git-commit"
  echo ""
  echo "   stack-name: AWS stack name. Must be uniquely identifiable"
  echo "   git-commit: Git Commit ID. Default: master"
}

if [ -z "${CF_STACK_NAME}" ]; then
  usage
  exit 1
fi
if [ -z "${GIT_COMMIT_ID}" ]; then
  GIT_COMMIT_ID="master"
fi

echo "Waiting for CF stack to come up ..."

aws cloudformation wait stack-create-complete \
  --stack-name ${CF_STACK_NAME}

# Get IP address of EC2 machine where tests can be executed
SCALEIO_NODE1=$(aws cloudformation describe-stack-resources --stack-name ${CF_STACK_NAME} | jq '.StackResources[] | select(.LogicalResourceId=="ScaleIONode1")' | jq -r .PhysicalResourceId)
SCALEIO_NODE2=$(aws cloudformation describe-stack-resources --stack-name ${CF_STACK_NAME} | jq '.StackResources[] | select(.LogicalResourceId=="ScaleIONode2")' | jq -r .PhysicalResourceId)
SCALEIO_NODE3=$(aws cloudformation describe-stack-resources --stack-name ${CF_STACK_NAME} | jq '.StackResources[] | select(.LogicalResourceId=="ScaleIONode3")' | jq -r .PhysicalResourceId)
#echo ${SCALEIO_NODE1}
#echo ${SCALEIO_NODE2}
#echo ${SCALEIO_NODE3}

SCALEIO_NODE1_FQDN=$(aws ec2 describe-instances --instance-ids ${SCALEIO_NODE1} | jq -r '.Reservations[0].Instances[0].PublicDnsName')
SCALEIO_NODE2_FQDN=$(aws ec2 describe-instances --instance-ids ${SCALEIO_NODE2} | jq -r '.Reservations[0].Instances[0].PublicDnsName')
SCALEIO_NODE3_FQDN=$(aws ec2 describe-instances --instance-ids ${SCALEIO_NODE3} | jq -r '.Reservations[0].Instances[0].PublicDnsName')
#echo ${SCALEIO_NODE1_FQDN}
#echo ${SCALEIO_NODE2_FQDN}
#echo ${SCALEIO_NODE3_FQDN}

ssh-keyscan -H ${SCALEIO_NODE1_FQDN} >> ~/.ssh/known_hosts
ssh-keyscan -H ${SCALEIO_NODE2_FQDN} >> ~/.ssh/known_hosts
ssh-keyscan -H ${SCALEIO_NODE3_FQDN} >> ~/.ssh/known_hosts

rm -f $(dirname $0)/output1.txt
rm -f $(dirname $0)/output2.txt
rm -f $(dirname $0)/output3.txt

# wait for scaleio-gateway to be running
READY1=$(ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE1_FQDN "service scaleio-gateway status 2>&1 | grep running")
while [ "$READY1" != "" ];
do
  sleep 1
  READY1=$(ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE1_FQDN "service scaleio-gateway status 2>&1 | grep running")
done
READY2=$(ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE2_FQDN "service scaleio-gateway status 2>&1 | grep running")
while [ "$READY2" != "" ];
do
  sleep 1
  READY2=$(ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE2_FQDN "service scaleio-gateway status 2>&1 | grep running")
done
READY3=$(ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE3_FQDN "service scaleio-gateway status 2>&1 | grep running")
while [ "$READY3" != "" ];
do
  sleep 1
  READY3=$(ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE3_FQDN "service scaleio-gateway status 2>&1 | grep running")
done

# ssh key to use
SSH_KEY_FILE=$(cat ./keyfile)

# Copy REX-Ray build to EC2 instance
scp -i ~/.ssh/${SSH_KEY_FILE} ./config.yml $CF_EC2_USER@$SCALEIO_NODE1_FQDN:/tmp
scp -i ~/.ssh/${SSH_KEY_FILE} ./tests1.sh $CF_EC2_USER@$SCALEIO_NODE1_FQDN:/tmp
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE1_FQDN "sudo chmod +x /tmp/tests1.sh"
if [ -f $GIT_COMMIT_ID ]; then
  scp -i ~/.ssh/${SSH_KEY_FILE} $GIT_COMMIT_ID $CF_EC2_USER@$SCALEIO_NODE1_FQDN:/tmp
  ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE1_FQDN "sudo chmod +x /tmp/rexray"
fi

scp -i ~/.ssh/${SSH_KEY_FILE} ./config.yml $CF_EC2_USER@$SCALEIO_NODE2_FQDN:/tmp
scp -i ~/.ssh/${SSH_KEY_FILE} ./tests2a.sh $CF_EC2_USER@$SCALEIO_NODE2_FQDN:/tmp
scp -i ~/.ssh/${SSH_KEY_FILE} ./tests2b.sh $CF_EC2_USER@$SCALEIO_NODE2_FQDN:/tmp
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE2_FQDN "sudo chmod +x /tmp/tests2a.sh"
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE2_FQDN "sudo chmod +x /tmp/tests2b.sh"
if [ -f $GIT_COMMIT_ID ]; then
  scp -i ~/.ssh/${SSH_KEY_FILE} $GIT_COMMIT_ID $CF_EC2_USER@$SCALEIO_NODE2_FQDN:/tmp
  ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE2_FQDN "sudo chmod +x /tmp/rexray"
fi

scp -i ~/.ssh/${SSH_KEY_FILE} ./config.yml $CF_EC2_USER@$SCALEIO_NODE3_FQDN:/tmp
scp -i ~/.ssh/${SSH_KEY_FILE} ./tests3.sh $CF_EC2_USER@$SCALEIO_NODE3_FQDN:/tmp
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE3_FQDN "sudo chmod +x /tmp/tests3.sh"
if [ -f $GIT_COMMIT_ID ]; then
  scp -i ~/.ssh/${SSH_KEY_FILE} $GIT_COMMIT_ID $CF_EC2_USER@$SCALEIO_NODE3_FQDN:/tmp
  ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE3_FQDN "sudo chmod +x /tmp/rexray"
fi

if [ -f $GIT_COMMIT_ID ]; then
  GIT_COMMIT_ID="/tmp/rexray"
fi

echo "Executing Tests!"

# Run tests... note the use of "bash --login -c". this forces a load of the user's
# env variables while cause the make command to fail.
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE1_FQDN "bash --login -c '/tmp/tests1.sh $CF_STACK_NAME $GIT_COMMIT_ID' 2>&1"
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE2_FQDN "bash --login -c '/tmp/tests2a.sh $CF_STACK_NAME $GIT_COMMIT_ID' 2>&1"
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE2_FQDN "bash --login -c '/tmp/tests2b.sh $CF_STACK_NAME $GIT_COMMIT_ID' 2>&1"
ssh -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE3_FQDN "bash --login -c '/tmp/tests3.sh $CF_STACK_NAME $GIT_COMMIT_ID' 2>&1"

# Copy test coverage results
scp -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE1_FQDN:/tmp/output.txt $(dirname $0)/output1.txt
scp -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE2_FQDN:/tmp/output.txt $(dirname $0)/output2.txt
scp -i ~/.ssh/${SSH_KEY_FILE} $CF_EC2_USER@$SCALEIO_NODE3_FQDN:/tmp/output.txt $(dirname $0)/output3.txt

echo "Tests passed and coverge results are available in the output files"
echo "Node 1 Results"
cat $(dirname $0)/output1.txt
echo " "
echo "Node 2 Results"
cat $(dirname $0)/output2.txt
echo " "
echo "Node 3 Results"
cat $(dirname $0)/output3.txt
