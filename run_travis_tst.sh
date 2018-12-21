#!/bin/sh

# ANGI TODO - remove later
pwd
ls -al
env

echo "zypper in ..."
zypper -n --gpg-auto-import-keys ref && zypper -n --gpg-auto-import-keys in systemd
systemctl start dbus
zypper -n --gpg-auto-import-keys ref && zypper -n --gpg-auto-import-keys in go1.7 go cpupower uuidd git

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

# ANGI TODO - remove later
echo "check system environment for tests"
pwd
ls -al
echo "available system parameter:"
sysctl -A
echo "check cpu governer"
ls /sys/devices/system/cpu/cpu0/cpufreq
cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_available_governors
echo "check /dev/shm"
mount | grep /dev/shm
df -h /dev/shm
df /dev/shm
echo "services"
systemctl
ls /etc/systemd
ls /etc/systemd/logind.conf.d
cat /etc/systemd/logind.conf.d/sap.conf

echo "go environment:"
go env
go version
echo "install goveralls"
go get github.com/mattn/goveralls

cd saptune
echo "run go tests"
#sudo -E env "PATH=$PATH" go test -v -cover ./...
go test -v -cover ./...
go build
echo "run goveralls"
$GOPATH/bin/goveralls -service=travis-ci
ls -l /root/.coverprofile

