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

FIRST_VOLUME=$CF_STACK_NAME"_4a"
SECOND_VOLUME=$CF_STACK_NAME"_4b"

# Clean Up the Env
sudo rm -f /tmp/output.txt
sudo rm -f /tmp/finished.txt
sudo mkdir -p /usr/bin
sudo mkdir -p /etc/rexray

if [ -f $GIT_LIBSTORAGE_COMMIT_ID ]; then
  echo "Using pre-built REX-Ray"

  sudo cp -f $GIT_LIBSTORAGE_COMMIT_ID /usr/bin
else
  echo "Building REX-Ray"
  echo "Building using libstorage commit: $GIT_LIBSTORAGE_COMMIT_ID"
  echo "Building using rexray commit: $GIT_REXRAY_COMMIT_ID"

  # Build REX-Ray
  rm -rf $HOME/.glide
  mkdir -p $HOME/go/src/github.com/codedellemc
  cd $HOME/go/src/github.com/codedellemc
  git clone https://github.com/codedellemc/libstorage.git
  cd $HOME/go/src/github.com/codedellemc/libstorage
  git checkout $GIT_LIBSTORAGE_COMMIT_ID
  cd $HOME/go/src/github.com/codedellemc
  git clone https://github.com/codedellemc/rexray.git
  cd $HOME/go/src/github.com/codedellemc/rexray
  sed -e "s/.*# libstorage-version/    ref:     $GIT_LIBSTORAGE_COMMIT_ID/g" -i glide.yaml
  sed -e $"s|.*# libstorage-repo|    repo:    file://$HOME/go/src/github.com/codedellemc/libstorage\n    vcs:     git|g" -i glide.yaml
  git checkout $GIT_REXRAY_COMMIT_ID
  rm -f glide.lock
  make deps
  make -j build-libstorage
  make build

  # Install REX-Ray
  sudo cp -f $HOME/go/bin/rexray /usr/bin
fi

sudo cp -f /tmp/config.yml /etc/rexray
sudo /usr/bin/rexray install
sudo service rexray restart

sleep 5

# Run the tests....
TEST1=$(sudo rexray volume create $FIRST_VOLUME --size 16 | grep "$FIRST_VOLUME" | awk '{print $2}')
if [ "$TEST1" == "$FIRST_VOLUME" ]; then
  printf "1:rexray volume create:PASS\n" >> /tmp/output.txt
else
  printf "1:rexray volume create:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST1" >> /tmp/output.txt
fi

TEST2=$(sudo rexray volume ls | grep "$FIRST_VOLUME" | awk '{print $2}')
if [ "$TEST2" == "$FIRST_VOLUME" ]; then
  printf "2:rexray volume ls:PASS\n" >> /tmp/output.txt
else
  printf "2:rexray volume ls:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST2" >> /tmp/output.txt
fi

TEST3=$(sudo rexray volume ls $FIRST_VOLUME --format=json | jq '.[0].name' | sed -e 's|["'\'']||g')
if [ "$TEST3" == "$FIRST_VOLUME" ]; then
  printf "3:rexray volume ls json:PASS\n" >> /tmp/output.txt
else
  printf "3:rexray volume ls json:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST3" >> /tmp/output.txt
fi

TEST4=$(sudo rexray volume ls $FIRST_VOLUME --format=jsonp | jq '.[0].name' | sed -e 's|["'\'']||g')
if [ "$TEST4" == "$FIRST_VOLUME" ]; then
  printf "4:rexray volume ls jsonp:PASS\n" >> /tmp/output.txt
else
  printf "4:rexray volume ls jsonp:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST4" >> /tmp/output.txt
fi

TEST5=$(sudo rexray volume attach $FIRST_VOLUME | grep "$FIRST_VOLUME" | awk '{print $3}')
if [ "$TEST5" == "attached" ]; then
  printf "5:rexray volume attach:PASS\n" >> /tmp/output.txt
else
  printf "5:rexray volume attach:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST5" >> /tmp/output.txt
fi

TEST6=$(sudo rexray volume mount $FIRST_VOLUME | grep "$FIRST_VOLUME" | awk '{print $5}')
if [ "$TEST6" == "/var/lib/libstorage/volumes/$FIRST_VOLUME/data" ]; then
  printf "6:rexray volume mount:PASS\n" >> /tmp/output.txt
else
  printf "6:rexray volume mount:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST6" >> /tmp/output.txt
fi

TEST7=$(sudo rexray volume ls $FIRST_VOLUME --format=jsonp | jq '.[0].attachments[0].instanceID.driver' | sed -e 's|["'\'']||g')
if [ "$TEST7" == "scaleio" ]; then
  printf "7:rexray volume ls jsonp:PASS\n" >> /tmp/output.txt
else
  printf "7:rexray volume ls jsonp:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST7" >> /tmp/output.txt
fi

TEST8=$(sudo rexray volume unmount $FIRST_VOLUME | grep "$FIRST_VOLUME" | awk '{print $2}')
if [ "$TEST8" == "$FIRST_VOLUME" ]; then
  printf "8:rexray volume unmount:PASS\n" >> /tmp/output.txt
else
  printf "8:rexray volume unmount:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST8" >> /tmp/output.txt
fi

TEST9=$(sudo rexray volume attach $FIRST_VOLUME | grep "$FIRST_VOLUME" | awk '{print $3}')
if [ "$TEST9" == "attached" ]; then
  printf "9:rexray volume attach:PASS\n" >> /tmp/output.txt
else
  printf "9:rexray volume attach:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST9" >> /tmp/output.txt
fi

TEST10=$(sudo rexray volume detach $FIRST_VOLUME | grep "$FIRST_VOLUME" | awk '{print $3}')
if [ "$TEST10" == "available" ]; then
  printf "10:rexray volume detach:PASS\n" >> /tmp/output.txt
else
  printf "10:rexray volume detach:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST10" >> /tmp/output.txt
fi

TEST11=$(sudo rexray volume rm $FIRST_VOLUME)
if [ "$TEST11" == "$FIRST_VOLUME" ]; then
  printf "11:rexray volume rm:PASS\n" >> /tmp/output.txt
else
  printf "11:rexray volume rm:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST11" >> /tmp/output.txt
fi

for i in `seq 1 10`;
do
  TEST12=$(sudo rexray volume ls | grep "$FIRST_VOLUME" | awk '{print $2}')
  if [ "$TEST12" == "" ]; then
    break
  fi
  sleep 1
done
if [ "$TEST12" == "" ]; then
  printf "12:rexray volume ls:PASS\n" >> /tmp/output.txt
else
  printf "12:rexray volume ls:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST12" >> /tmp/output.txt
fi

TEST13=$(sudo docker volume create --driver rexray --name $SECOND_VOLUME --opt size=16)
if [ "$TEST13" == "$SECOND_VOLUME" ]; then
  printf "13:docker volume create:PASS\n" >> /tmp/output.txt
else
  printf "13:docker volume create:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST13" >> /tmp/output.txt
fi

TEST14=$(sudo docker run -d --volume-driver=rexray -v $SECOND_VOLUME:/tmp dvonthenen/demo-boot)
if [ "$TEST14" != "" ]; then
  printf "14:docker run mount:PASS\n" >> /tmp/output.txt
else
  printf "14:docker run mount:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST14" >> /tmp/output.txt
fi

TEST15=$(sudo docker volume inspect $SECOND_VOLUME | jq '.[0].Mountpoint' | sed -e 's|["'\'']||g')
if [ "$TEST15" == "/var/lib/libstorage/volumes/$SECOND_VOLUME/data" ]; then
  printf "15:docker volume inspect:PASS\n" >> /tmp/output.txt
else
  printf "15:docker volume inspect:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST15" >> /tmp/output.txt
fi

TEST16=$(sudo docker ps | grep 'dvonthenen/demo-boot' | awk '{print $1}' | xargs sudo docker stop)
if [ "$TEST16" != "" ]; then
  printf "16:docker volume unmount:PASS\n" >> /tmp/output.txt
else
  printf "16:docker volume unmount:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST16" >> /tmp/output.txt
fi

TEST17=$(sudo docker volume inspect $SECOND_VOLUME | jq '.[0].Mountpoint' | sed -e 's|["'\'']||g')
if [ "$TEST17" == "/" ]; then
  printf "17:docker volume inspect:PASS\n" >> /tmp/output.txt
else
  printf "17:docker volume inspect:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST17" >> /tmp/output.txt
fi

#Remove old docker Instances
sudo docker rm $(sudo docker ps -a | grep Exited | awk '{print $1}')

TEST18=$(sudo docker volume rm $SECOND_VOLUME)
if [ "$TEST18" == "$SECOND_VOLUME" ]; then
  printf "18:docker volume rm:PASS\n" >> /tmp/output.txt
else
  printf "18:docker volume rm:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST18" >> /tmp/output.txt
fi

for i in `seq 1 10`;
do
  TEST19=$(sudo docker volume ls | grep "$SECOND_VOLUME" | awk '{print $2}')
  if [ "$TEST19" == "" ]; then
    break
  fi
  sleep 1
done
if [ "$TEST19" == "" ]; then
  printf "19:docker volume ls:PASS\n" >> /tmp/output.txt
else
  printf "19:docker volume ls:FAILED\n" >> /tmp/output.txt
fi
if [ "$DEBUG" == "true" ]; then
  printf "#%s\n" "$TEST19" >> /tmp/output.txt
fi

echo "finished" > /tmp/finished.txt
