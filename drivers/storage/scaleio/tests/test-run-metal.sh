#!/usr/bin/env bash

# This script runs tests

# set -e
# set -x

: ${LINUX_USER:="rexray"}

GIT_COMMIT_ID=${GIT_COMMIT_ID:-$1}
NODE1_FQDN=${NODE1_FQDN:-$2}
NODE2_FQDN=${NODE2_FQDN:-$3}
NODE3_FQDN=${NODE3_FQDN:-$4}
#NODE4_FQDN=${NODE4_FQDN:-$5}

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
  echo "Usage: ${0} git-commit node1_fqdn node2_fqdn node3_fqdn, node4_fqdn"
  echo ""
  echo "   git-commit: Git Commit ID. Default: master"
}

if [ -z "${GIT_COMMIT_ID}" ]; then
  GIT_COMMIT_ID="master"
fi
if [ -z "${NODE1_FQDN}" ]; then
  usage
  exit 1
fi
if [ -z "${NODE2_FQDN}" ]; then
  usage
  exit 1
fi
if [ -z "${NODE3_FQDN}" ]; then
  usage
  exit 1
fi
#if [ -z "${NODE4_FQDN}" ]; then
#  usage
#  exit 1
#fi

ssh-keyscan -H ${NODE1_FQDN} >> ~/.ssh/known_hosts
ssh-keyscan -H ${NODE2_FQDN} >> ~/.ssh/known_hosts
ssh-keyscan -H ${NODE3_FQDN} >> ~/.ssh/known_hosts
#ssh-keyscan -H ${NODE4_FQDN} >> ~/.ssh/known_hosts

rm -f $(dirname $0)/output1.txt
rm -f $(dirname $0)/output2.txt
rm -f $(dirname $0)/output3.txt
#rm -f $(dirname $0)/output4.txt

# Copy REX-Ray build to EC2 instance
scp ./config.yml $LINUX_USER@$NODE1_FQDN:/tmp
scp ./tests1.sh $LINUX_USER@$NODE1_FQDN:/tmp
ssh $LINUX_USER@$NODE1_FQDN "sudo chmod +x /tmp/tests1.sh"

scp ./config.yml $LINUX_USER@$NODE2_FQDN:/tmp
scp ./tests2a.sh $LINUX_USER@$NODE2_FQDN:/tmp
scp ./tests2b.sh $LINUX_USER@$NODE2_FQDN:/tmp
ssh $LINUX_USER@$NODE2_FQDN "sudo chmod +x /tmp/tests2a.sh"
ssh $LINUX_USER@$NODE2_FQDN "sudo chmod +x /tmp/tests2b.sh"

scp ./config.yml $LINUX_USER@$NODE3_FQDN:/tmp
scp ./tests3.sh $LINUX_USER@$NODE3_FQDN:/tmp
ssh $LINUX_USER@$NODE3_FQDN "sudo chmod +x /tmp/tests3.sh"

#scp ./config.yml $LINUX_USER@$NODE4_FQDN:/tmp
#scp ./tests4.sh $LINUX_USER@$NODE4_FQDN:/tmp
#ssh $LINUX_USER@$NODE4_FQDN "sudo chmod +x /tmp/tests4.sh"

echo "Executing Tests!"

# Run tests... note the use of "bash --login -c". this forces a load of the user's
# env variables while cause the make command to fail.
ssh $LINUX_USER@$NODE1_FQDN "bash --login -c '/tmp/tests1.sh $GIT_COMMIT_ID' 2>&1" &
ssh $LINUX_USER@$NODE2_FQDN "bash --login -c '/tmp/tests2a.sh $GIT_COMMIT_ID' 2>&1" &
ssh $LINUX_USER@$NODE3_FQDN "bash --login -c '/tmp/tests3.sh $GIT_COMMIT_ID' 2>&1" &
#ssh $LINUX_USER@$NODE4_FQDN "bash --login -c '/tmp/tests4.sh $GIT_COMMIT_ID' 2>&1" &

FINISHED1=$(ssh $LINUX_USER@$NODE1_FQDN "cat /tmp/finished.txt 2>&1")
while [ "$FINISHED1" != "finished" ];
do
  sleep 1
  FINISHED1=$(ssh $LINUX_USER@$NODE1_FQDN "cat /tmp/finished.txt 2>&1")
done

# execute the preemption portion of the script... note that node 1 and node 3
# should finish before node 2. see comment below.
ssh $LINUX_USER@$NODE2_FQDN "bash --login -c '/tmp/tests2b.sh $GIT_COMMIT_ID' 2>&1" &

FINISHED3=$(ssh $LINUX_USER@$NODE3_FQDN "cat /tmp/finished.txt 2>&1")
while [ "$FINISHED3" != "finished" ];
do
  sleep 1
  FINISHED3=$(ssh $LINUX_USER@$NODE3_FQDN "cat /tmp/finished.txt 2>&1")
done

#FINISHED4=$(ssh $LINUX_USER@$NODE4_FQDN "cat /tmp/finished.txt 2>&1")
#while [ "$FINISHED4" != "finished" ];
#do
#  sleep 1
#  FINISHED4=$(ssh $LINUX_USER@$NODE4_FQDN "cat /tmp/finished.txt 2>&1")
#done

# this node should finish last because of the preemption test. hence this order
# (Node 1, 3, 2) for checking for finished
FINISHED2=$(ssh $LINUX_USER@$NODE2_FQDN "cat /tmp/finished.txt 2>&1")
while [ "$FINISHED2" != "finished" ];
do
  sleep 1
  FINISHED2=$(ssh $LINUX_USER@$NODE2_FQDN "cat /tmp/finished.txt 2>&1")
done

# Copy test coverage results
scp $LINUX_USER@$NODE1_FQDN:/tmp/output.txt $(dirname $0)/output1.txt
scp $LINUX_USER@$NODE2_FQDN:/tmp/output.txt $(dirname $0)/output2.txt
scp $LINUX_USER@$NODE3_FQDN:/tmp/output.txt $(dirname $0)/output3.txt
#scp $LINUX_USER@$NODE4_FQDN:/tmp/output.txt $(dirname $0)/output4.txt

echo "Tests passed and coverge results are available in the output files"
echo "Node 1 Results"
cat $(dirname $0)/output1.txt
echo " "
echo "Node 2 Results"
cat $(dirname $0)/output2.txt
echo " "
echo "Node 3 Results"
cat $(dirname $0)/output3.txt
#echo " "
#echo "Node 4 Results"
#cat $(dirname $0)/output4.txt
