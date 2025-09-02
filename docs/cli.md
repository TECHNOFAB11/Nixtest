# CLI

```sh title="nix run .#nixtests:run -- --help"
Usage of nixtest:
      --junit string          Path to generate JUNIT report to, leave empty to disable
      --no-color              Disable coloring
      --impure                Don\'t unset all env vars before running script tests
  -s, --skip string           Regular expression to skip tests (e.g., 'test-.*|.*-b')
      --snapshot-dir string   Directory where snapshots are stored (default "./snapshots")
  -f, --tests string          Path to JSON file containing tests (required)
  -u, --update-snapshots      Update all snapshots
  -w, --workers int           Amount of tests to run in parallel (default 4)
```
