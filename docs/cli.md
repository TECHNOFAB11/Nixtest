# CLI

```sh title="nix run .#nixtests:run -- --help"
Usage of nixtest:
      --junit string          Path to generate JUNIT report to, leave empty to disable
      --pure                  Unset all env vars before running script tests
      --skip string           Regular expression to skip (e.g., 'test-.*|.*-b')
      --snapshot-dir string   Directory where snapshots are stored (default "./snapshots")
      --tests string          Path to JSON file containing tests
      --update-snapshots      Update all snapshots
      --workers int           Amount of tests to run in parallel (default 4)
```
