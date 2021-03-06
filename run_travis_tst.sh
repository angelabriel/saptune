#!/bin/sh

# this tests do not need any HA/cluster stuff, so remove the repo
echo "zypper remove unneeded repo"
zypper rr network:ha-clustering:Factory

echo "zypper in ..."
#/bin/systemctl start dbus -> does not work any longer
# additional libs needed to get 'tuned' working
zypper -n --gpg-auto-import-keys ref && zypper -n --gpg-auto-import-keys in glib2 glib2-tools libgio-2_0-0 libglib-2_0-0 libgmodule-2_0-0 libgobject-2_0-0 go1.10 go rpcbind cpupower uuidd polkit tuned sysstat

# dbus can not be started directly, only by dependency - so start 'tuned' instead
/bin/systemctl start tuned
systemctl --no-pager status
# try to resolve systemd status 'degraded'
systemctl reset-failed
systemctl --no-pager status

echo "PATH is $PATH, GOPATH is $GOPATH, TRAVIS_HOME is $TRAVIS_HOME"

export TRAVIS_HOME=/home/travis
mkdir -p ${TRAVIS_HOME}/gopath/src/github.com/SUSE
cd ${TRAVIS_HOME}/gopath/src/github.com/SUSE
if [ ! -f saptune ]; then
	ln -s /app saptune
fi
export GO111MODULE=off
export GOPATH=${TRAVIS_HOME}/gopath
export PATH=${TRAVIS_HOME}/gopath/bin:$PATH
export TRAVIS_BUILD_DIR=${TRAVIS_HOME}/gopath/src/github.com/SUSE/saptune

echo "PATH is $PATH, GOPATH is $GOPATH, TRAVIS_HOME is $TRAVIS_HOME"
echo "ls -l /etc/saptune/*"
ls -l /etc/saptune/*

mkdir -p /etc/saptune/override
mkdir -p /var/lib/saptune/working
if [ ! -f /usr/share/saptune/solutions ]; then
	ln -s /app/testdata/saptune-test-solutions /var/lib/saptune/working/solutions
fi

echo "go environment:"
go env
go version

cd saptune
pwd
ls -al

# to get TasksMax settings work, needs a user login session
echo "start nobody login session in background"
su --login nobody -c "sleep 4m" &
sleep 10
ps -ef
loginctl --no-pager

echo "exchange /etc/os-release"
cp /etc/os-release /etc/os-release_OrG

# for some sysctl tests
echo "vm.pagecache_limit_ignore_dirty = 1" > /etc/sysctl.d/saptune_test.conf
echo "vm.pagecache_limit_ignore_dirty = 1" > /etc/sysctl.d/saptune_test2.conf

echo "run go tests"
go test -v -coverprofile=c.out -cover ./...
exitErr=$?
go build
ps -ef
pkill -P $$
exit $exitErr
