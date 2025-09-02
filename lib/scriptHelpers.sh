output=
exit_code=

function assert() {
  test $1 || { echo "Assertion '$1' failed: $2" >&2; exit 1; }
}
function assert_eq() {
  assert "$1 -eq $2" "$3"
}
function assert_not_eq() {
  assert "$1 -ne $2" "$3"
}
function assert_contains() {
  echo "$1" | grep -q -- "$2" || {
    echo "Assertion failed: $3. $1 does not contain $2" >&2;
    exit 1;
  }
}
function assert_not_contains() {
  echo "$1" | grep -q -- "$2" && {
    echo "Assertion failed: $3. $1 does contain $2" >&2;
    exit 1;
  }
}
function assert_file_contains() {
  grep -q -- "$2" $1 || {
    echo "Assertion failed: $3. $1 does not contain $2" >&2;
    exit 1;
  }
}
function assert_file_not_contains() {
  grep -q -- "$2" $1 && {
    echo "Assertion failed: $3. $1 does contain $2" >&2;
    exit 1;
  }
}

function run() {
  output=$($@ 2>&1)
  exit_code=$?
}
