# Usage

## Flake Module

```nix
{
  inputs.nixtest.url = "gitlab:TECHNOFAB/nixtest?dir=lib";
  # ... mkFlake ...
  imports = [
    inputs.nixtest.flakeModule
  ];
  
  # perSystem
    nixtest = {
      # regex of tests to skip. Can also be passed as CLI arg
      skip = "";
      suites = {
        "Suite A" = {
          # pos shows the file the test was declared in in the summary and 
          # junit report by setting it on the suite, all tests inherit this pos
          pos = __curPos; 
          tests = [
            # define tests here (see below)
          ];
        };
        "Suite B" = {
          # etc.
        };
      };
    };
  # ...
}
```

## Library

You can also integrate nixtest in your own workflow by using the lib functions directly.
Check out `flakeModule.nix` to see how it's used there.

<!-- TODO: more detailed? -->

## Define Tests

There are currently 3 types of tests:

- `snapshot` -> snapshot testing, only needs `actual` and compares that to the snapshot
- `unit` -> equality checking, needs `expected` and `actual` or `actualDrv`
- `script` -> shell script test, needs `script`

Examples:

```nix
[
  {
    name = "unit-test"; # required
    type = "unit";  # default is unit
    expected = 1;
    actual = 1;
  }
  {
    name = "snapshot-test";
    type = "snapshot";
    # snapshot tests use snapshot files (stored by default in ./snapshots/)
    # and compare the "actual" value below with these files
    actual = 1;
  }
  {
    name = "snapshot-derivation-test";
    type = "snapshot";
    # instead of passing a nix expression, we can also use a derivation to do
    # more complex stuff. Will only be built when running the test (+ included
    # in the test time).
    actualDrv = pkgs.runCommand "test-snapshot" {} ''
      echo '"snapshot drv"' > $out
    '';
  }
  {
    name = "script-test";
    type = "script";
    script = 
    # there are two modes, "default"/"impure" and "pure"
    # in impure mode all env variables etc. from your current session are kept
    # and are available to the test
    # to make it more reproducible and cleaner, use --pure to switch to pure
    # mode which will unset all env variables before running the test. That 
    # requires you to set PATH yourself then:
    ''
      export PATH="${lib.makeBinPath [pkgs.gnugrep]}"
      grep -q "test" ${builtins.toFile "test" "test"}
    '';
  }
  {
    name = "pretty-test";
    # by default it uses json to serialize and compare the values. Derivations
    # and functions don't really work that way though, so you can also use
    # "pretty" to use lib.generators.pretty
    format = "pretty";
    # you can also set the pos here
    pos = __curPos;
    expected = pkgs.hello;
    actual = pkgs.hello;
  }
]
```
