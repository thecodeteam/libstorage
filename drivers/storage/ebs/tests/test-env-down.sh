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

CF_STACK_NAME=$(cat ./ebs-uniquename)
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

VOL1ANAME=$CF_STACK_NAME"_1a"
VOL1BNAME=$CF_STACK_NAME"_1b"
VOL2ANAME=$CF_STACK_NAME"_2a"
VOL2BNAME=$CF_STACK_NAME"_2b"
VOL3ANAME=$CF_STACK_NAME"_3a"
VOL3BNAME=$CF_STACK_NAME"_3b"

VOL1A=$(aws ec2 describe-volumes | jq '.Volumes[] | select(.Tags[]?.Value=="'${VOL1ANAME}'")' | jq -r .VolumeId)
VOL1B=$(aws ec2 describe-volumes | jq '.Volumes[] | select(.Tags[]?.Value=="'${VOL1BNAME}'")' | jq -r .VolumeId)
VOL2A=$(aws ec2 describe-volumes | jq '.Volumes[] | select(.Tags[]?.Value=="'${VOL2ANAME}'")' | jq -r .VolumeId)
VOL2B=$(aws ec2 describe-volumes | jq '.Volumes[] | select(.Tags[]?.Value=="'${VOL2BNAME}'")' | jq -r .VolumeId)
VOL3A=$(aws ec2 describe-volumes | jq '.Volumes[] | select(.Tags[]?.Value=="'${VOL3ANAME}'")' | jq -r .VolumeId)
VOL3B=$(aws ec2 describe-volumes | jq '.Volumes[] | select(.Tags[]?.Value=="'${VOL3BNAME}'")' | jq -r .VolumeId)
#echo $VOL1A
#echo $VOL1B
#echo $VOL2A
#echo $VOL2B
#echo $VOL3A
#echo $VOL3B

if [ "$VOL1A" != "" ]; then
  aws ec2 detach-volume --volume-id $VOL1A
  aws ec2 delete-volume --volume-id $VOL1A
fi
if [ "$VOL1B" != "" ]; then
  aws ec2 detach-volume --volume-id $VOL1B
  aws ec2 delete-volume --volume-id $VOL1B
fi
if [ "$VOL2A" != "" ]; then
  aws ec2 detach-volume --volume-id $VOL2A
  aws ec2 delete-volume --volume-id $VOL2A
fi
if [ "$VOL2B" != "" ]; then
  aws ec2 detach-volume --volume-id $VOL2B
  aws ec2 delete-volume --volume-id $VOL2B
fi
if [ "$VOL3A" != "" ]; then
  aws ec2 detach-volume --volume-id $VOL3A
  aws ec2 delete-volume --volume-id $VOL3A
fi
if [ "$VOL3B" != "" ]; then
  aws ec2 detach-volume --volume-id $VOL3B
  aws ec2 delete-volume --volume-id $VOL3B
fi

rm -f ./ebs-uniquename

echo "Stack has been deleted"
