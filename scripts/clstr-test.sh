#!/usr/bin/env bash

set -u

before_containers="$(docker ps -aq --filter label=io.clstr=true)"

cleanup() {
	test_status=$?
	trap - EXIT INT TERM

	after_containers="$(docker ps -aq --filter label=io.clstr=true)"

	for container in $after_containers; do
		if printf '%s\n' "$before_containers" | grep -qx "$container"; then
			continue
		fi

		echo "Cleaning up clstr container $container"
		docker stop -t 10 "$container" >/dev/null 2>&1 || true
		docker rm "$container" >/dev/null
	done

	exit "$test_status"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

clstr test "$@"
