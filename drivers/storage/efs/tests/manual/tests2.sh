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

FIRST_VOLUME=$CF_STACK_NAME"_1a"
SECOND_VOLUME=$CF_STACK_NAME"_1b"

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

echo "finished" > /tmp/finished.txt
