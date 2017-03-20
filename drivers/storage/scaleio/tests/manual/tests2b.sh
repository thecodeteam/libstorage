#!/usr/bin/env bash

# This script builds REX-Ray and then runs tests

DEBUG=false

CF_STACK_NAME=${CF_STACK_NAME:-$1}
GIT_LIBSTORAGE_COMMIT_ID=${GIT_LIBSTORAGE_COMMIT_ID:-$2}
GIT_REXRAY_COMMIT_ID=${GIT_REXRAY_COMMIT_ID:-$3}

if [ -z "${CF_STACK_NAME}" ]; then
  CF_STACK_NAME="default"
fi
if [ -z "${GIT_LIBSTORAGE_COMMIT_ID}" ]; then
  GIT_LIBSTORAGE_COMMIT_ID="master"
fi
if [ -z "${GIT_REXRAY_COMMIT_ID}" ]; then
  GIT_REXRAY_COMMIT_ID="master"
fi

PREEMPT_VOLUME=$CF_STACK_NAME"_1b"

# Run the tests....
TEST20=$(sudo docker run -d --volume-driver=rexray -v $PREEMPT_VOLUME:/tmp dvonthenen/demo-boot)
if [ "$TEST20" != "" ]; then
  printf "20:docker run mount preempt:PASS\n" >> /tmp/output.txt
else
  printf "20:docker run mount preempt:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST20" >> /tmp/output.txt
fi

TEST21=$(sudo docker volume inspect $PREEMPT_VOLUME | jq '.[0].Mountpoint' | sed -e 's|["'\'']||g')
if [ "$TEST21" == "/var/lib/libstorage/volumes/$PREEMPT_VOLUME/data" ]; then
  printf "21:docker volume inspect preempt:PASS\n" >> /tmp/output.txt
else
  printf "21:docker volume inspect preempt:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST21" >> /tmp/output.txt
fi

TEST22=$(sudo docker ps | grep 'dvonthenen/demo-boot' | awk '{print $1}' | xargs sudo docker stop)
if [ "$TEST22" != "" ]; then
  printf "22:docker volume unmount preempt:PASS\n" >> /tmp/output.txt
else
  printf "22:docker volume unmount preempt:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST22" >> /tmp/output.txt
fi

TEST23=$(sudo docker volume inspect $PREEMPT_VOLUME | jq '.[0].Mountpoint' | sed -e 's|["'\'']||g')
if [ "$TEST23" == "/" ]; then
  printf "23:docker volume inspect preempt:PASS\n" >> /tmp/output.txt
else
  printf "23:docker volume inspect preempt:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST23" >> /tmp/output.txt
fi

#Remove old docker Instances
sudo docker rm $(sudo docker ps -a | grep Exited | awk '{print $1}')

TEST24=$(sudo docker volume rm $PREEMPT_VOLUME)
if [ "$TEST24" == "$PREEMPT_VOLUME" ]; then
  printf "24:docker volume rm preempt:PASS\n" >> /tmp/output.txt
else
  printf "24:docker volume rm preempt:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST24" >> /tmp/output.txt
fi

for i in `seq 1 10`;
do
  TEST25=$(sudo docker volume ls | grep "$PREEMPT_VOLUME" | awk '{print $2}')
  if [ "$TEST25" == "" ]; then
    break
  fi
  sleep 1
done
if [ "$TEST25" == "" ]; then
  printf "25:docker volume ls preempt:PASS\n" >> /tmp/output.txt
else
  printf "25:docker volume ls preempt:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST25" >> /tmp/output.txt
fi

echo "finished" > /tmp/finished.txt
