#!/usr/bin/env bash

set -u

project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
script_under_test="$project_root/scripts/clstr-test.sh"
test_dir="$(mktemp -d)"
fake_bin="$test_dir/bin"
docker_log="$test_dir/docker.log"
clstr_log="$test_dir/clstr.log"
state_file="$test_dir/test-started"

cleanup_test() {
	rm -rf "$test_dir"
}
trap cleanup_test EXIT

mkdir -p "$fake_bin"

cat >"$fake_bin/docker" <<'EOF'
#!/usr/bin/env bash
if [[ "$1 $2" == "ps -aq" ]]; then
	echo "existing-clstr"
	if [[ -f "$DOCKER_STATE_FILE" ]]; then
		echo "new-running"
		echo "new-stopped"
	fi
	exit 0
fi

printf '%s\n' "$*" >>"$DOCKER_LOG"
EOF

cat >"$fake_bin/clstr" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$CLSTR_LOG"
touch "$DOCKER_STATE_FILE"
exit 7
EOF

chmod +x "$fake_bin/docker" "$fake_bin/clstr"

set +e
PATH="$fake_bin:$PATH" \
	DOCKER_LOG="$docker_log" \
	CLSTR_LOG="$clstr_log" \
	DOCKER_STATE_FILE="$state_file" \
	"$script_under_test" http-api
status=$?
set -e

if [[ "$status" -ne 7 ]]; then
	echo "expected clstr exit status 7, got $status" >&2
	exit 1
fi

if [[ "$(cat "$clstr_log")" != "test http-api" ]]; then
	echo "expected arguments to be forwarded to clstr test" >&2
	exit 1
fi

for container in new-running new-stopped; do
	if ! grep -qx "stop -t 10 $container" "$docker_log"; then
		echo "expected $container to be stopped" >&2
		exit 1
	fi
	if ! grep -qx "rm $container" "$docker_log"; then
		echo "expected $container to be removed" >&2
		exit 1
	fi
done

if grep -q "existing-clstr" "$docker_log"; then
	echo "existing clstr container must not be modified" >&2
	exit 1
fi

echo "clstr cleanup wrapper test passed"
