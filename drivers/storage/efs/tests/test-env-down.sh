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

CF_STACK_NAME=$(cat ./efs-uniquename)
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

FIRST_VOLUME=$CF_STACK_NAME"_1a"
SECOND_VOLUME=$CF_STACK_NAME"_1b"

EFSID1=$(aws efs describe-file-systems | jq '.FileSystems[] | select(.Name=="'$FIRST_VOLUME'")' | jq -r .FileSystemId)
#echo $EFSID1
if [ "$EFSID1" != "" ]; then
  MOUNTS1=$(aws efs describe-mount-targets --file-system-id $EFSID1 | jq -r .MountTargets[]?.MountTargetId)
  #echo $MOUNTS1

  while read -r line; do
    if [ "$line" == "" ]; then
      continue
    fi
    aws efs delete-mount-target --mount-target-id $line
  done <<< "$MOUNTS1"

  DELETE1=$(aws efs delete-file-system --file-system-id $EFSID1 2>&1 | grep FileSystemInUse)
  #echo $DELETE1
  while [ "$DELETE1" != "" ];
  do
    sleep 1
    DELETE1=$(aws efs delete-file-system --file-system-id $EFSID1 2>&1 | grep FileSystemInUse)
    #echo $DELETE1
  done
fi

EFSID2=$(aws efs describe-file-systems | jq '.FileSystems[] | select(.Name=="'$SECOND_VOLUME'")' | jq -r .FileSystemId)
#echo $EFSID2
if [ "$EFSID2" != "" ]; then
  MOUNTS2=$(aws efs describe-mount-targets --file-system-id $EFSID2 | jq -r .MountTargets[]?.MountTargetId)
  #echo $MOUNTS2

  while read -r line; do
    if [ "$line" == "" ]; then
      continue
    fi
    aws efs delete-mount-target --mount-target-id $line
  done <<< "$MOUNTS2"

  DELETE2=$(aws efs delete-file-system --file-system-id $EFSID2 2>&1 | grep FileSystemInUse)
  #echo $DELETE2
  while [ "$DELETE2" != "" ];
  do
    sleep 1
    DELETE2=$(aws efs delete-file-system --file-system-id $EFSID2 2>&1 | grep FileSystemInUse)
    #echo $DELETE2
  done
fi

rm -f ./efs-uniquename

echo "Stack has been deleted"
