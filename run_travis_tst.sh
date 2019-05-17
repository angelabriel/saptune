#!/bin/sh

pwd
ls -al
ls -l /sys/devices
ls -l /sys/devices/system/cpu
ls -l /sys/devices/system/cpu*/*

echo "zypper in ..."
#/bin/systemctl start dbus
zypper -n --gpg-auto-import-keys ref && zypper -n --gpg-auto-import-keys in go1.10 go cpupower uuidd polkit tuned
/bin/systemctl start tuned
systemctl status

echo "PATH is $PATH, GOPATH is $GOPATH, TRAVIS_HOME is $TRAVIS_HOME"

export TRAVIS_HOME=/home/travis
mkdir -p ${TRAVIS_HOME}/gopath/src/github.com/SUSE
cd ${TRAVIS_HOME}/gopath/src/github.com/SUSE
ln -s /app saptune
export GOPATH=${TRAVIS_HOME}/gopath
export PATH=${TRAVIS_HOME}/gopath/bin:$PATH
export TRAVIS_BUILD_DIR=${TRAVIS_HOME}/gopath/src/github.com/SUSE/saptune

mkdir -p /usr/share/saptune
ln -s /app/testdata/saptune-test-solutions /usr/share/saptune/solutions

echo "go environment:"
go env
go version

cd saptune
pwd
ls -al
echo "run go tests"
#go test -v -cover ./... -coverprofile=cover.out
go test -v -coverprofile=c.out -cover ./...
exitErr=$?
go build
exit $exitErr
