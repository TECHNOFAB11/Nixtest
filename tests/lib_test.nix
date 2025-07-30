{
  pkgs,
  ntlib,
  ...
}: {
  suites."Lib Tests" = {
    pos = __curPos;
    tests = [
      {
        name = "autodiscovery";
        type = "script";
        script = let
          actual = ntlib.helpers.toPrettyFile (ntlib.autodiscover {
            dir = ./fixtures;
          });
          # tests if strings with store path context work
          actualDirString = ntlib.helpers.toPrettyFile (ntlib.autodiscover {
            dir = "${./fixtures}";
          });
        in
          # sh
          ''
            ${ntlib.helpers.path [pkgs.gnugrep]}
            ${ntlib.helpers.scriptHelpers}
            assert_file_contains ${actual} "sample_test.nix" "should find sample_test.nix"
            assert_file_contains ${actual} "base = \"/nix/store/.*-source/tests/fixtures/\"" "should set base to fixtures dir"

            assert_file_contains ${actualDirString} "sample_test.nix" "should find sample_test.nix"
            assert_file_contains ${actualDirString} "base = \"/nix/store/.*-fixtures/\"" "should set base to fixtures dir"
          '';
      }
      {
        name = "binary";
        type = "script";
        script = let
          binary =
            (ntlib.mkBinary {
              nixtests = "stub";
              extraParams = "--pure";
            })
            + "/bin/nixtests:run";
        in
          # sh
          ''
            ${ntlib.helpers.path [pkgs.gnugrep]}
            ${ntlib.helpers.scriptHelpers}
            assert_file_contains ${binary} "nixtest" "should contain nixtest"
            assert_file_contains ${binary} "--pure" "should contain --pure arg"
            assert_file_contains ${binary} "--tests=stub" "should contain --tests arg"

            run "${binary} --help"
            assert_eq $exit_code 0 "should exit 0"
            assert_contains "$output" "Usage of nixtest" "should show help"

            run "${binary}"
            assert_eq $exit_code 1 "should exit 1"
            assert_contains "$output" "Tests file does not exist"
          '';
      }
      {
        name = "full run with fixtures";
        type = "script";
        script = let
          binary =
            (ntlib.mkNixtest {
              modules = ntlib.autodiscover {dir = ./fixtures;};
              args = {inherit pkgs;};
            })
            + "/bin/nixtests:run";
        in
          # sh
          ''
            ${ntlib.helpers.path [pkgs.gnugrep pkgs.mktemp]}
            ${ntlib.helpers.scriptHelpers}

            TMPDIR=$(tmpdir)
            # start without nix & env binaries to expect errors
            run "${binary} --pure --junit=$TMPDIR/junit.xml"
            assert "$exit_code -eq 2" "should exit 2"
            assert "-f $TMPDIR/junit.xml" "should create junit.xml"
            assert_contains "$output" "executable file not found" "nix should not be found in pure mode"

            # now add required deps
            ${ntlib.helpers.pathAdd [pkgs.nix pkgs.coreutils]}
            run "${binary} --pure --junit=$TMPDIR/junit2.xml"
            assert "$exit_code -eq 2" "should exit 2"
            assert "-f $TMPDIR/junit2.xml" "should create junit2.xml"
            assert_not_contains "$output" "executable file not found" "nix should now exist"
            assert_contains "$output" "suite-one" "should contain suite-one"
            assert_contains "$output" "8/11 (1 SKIPPED)" "should be 8/11 total"
            assert_contains "$output" "ERROR" "should contain an error"
            assert_contains "$output" "SKIP" "should contain a skip"
          '';
      }
    ];
  };
}
