#!/bin/bash

BASEDIR="$(dirname "$(perl -e 'use Cwd qw/abs_path/; print abs_path($ARGV[0]);' "$0")")"

cd "${BASEDIR}"/..

set -e

#===============================================================================
function __run_coverage {
#===============================================================================
	if [ -d ./sbin/coverage ]; then
		rm -r ./sbin/coverage
	fi
	mkdir -p ./sbin/coverage

	set +e

	echo "mode: count" > ./sbin/coverage.out

	export env=local-tests

	local paths_to_packages
	if [ -z "$1" ]; then
		paths_to_packages=$(glide nv)
	else
		for _pkg in $@; do
			paths_to_packages="$paths_to_packages ./$_pkg/..."
		done
	fi

	for pkg in $(go list $paths_to_packages); do
		_pkg=$(echo $pkg | sed 's/\//_/g')
		if ! go test -coverprofile ./sbin/coverage/$_pkg.tmp $pkg; then
			echo "Failed to run tests!"
		fi
		if [ -f ./sbin/coverage/$_pkg.tmp ]; then
			tail -n +2 ./sbin/coverage/$_pkg.tmp > ./sbin/coverage/$_pkg
			rm ./sbin/coverage/$_pkg.tmp
		fi
	done

	for _file in ./sbin/coverage/*; do
		cat $_file >> ./sbin/coverage.out
	done

	go tool cover -html=./sbin/coverage.out -o ./sbin/coverage.html

	set -e
}

case "$1" in
	'test')
		go test -race -v $(glide nv)
		# replace with the line below (in the Makefile) when using Go 1.9
		#go test -v ./...
	;;

	'coverage')
		shift
		__run_coverage $@
	;;
esac
