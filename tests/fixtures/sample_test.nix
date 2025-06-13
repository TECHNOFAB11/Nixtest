{
  lib,
  pkgs,
  ...
}: {
  skip = "skip.*d";
  suites = {
    "suite-one" = {
      # required to figure out file and line, but optional
      pos = __curPos;
      tests = [
        {
          name = "test-one";
          expected = 1;
          actual = 1;
        }
        {
          name = "fail";
          expected = 0;
          actual = "meow";
        }
        {
          name = "snapshot-test";
          type = "snapshot";
          actual = "test";
        }
        {
          name = "test-snapshot-drv";
          type = "snapshot";
          actualDrv = pkgs.runCommand "test-snapshot" {} ''
            echo '"snapshot drv"' > $out
          '';
        }
        {
          name = "test-error-drv";
          expected = null;
          actualDrv = pkgs.runCommand "test-error-drv" {} ''
            echo "This works, but its better to just write 'fail' to \$out and expect 'success' or sth."
            exit 1
          '';
        }
        {
          name = "test-script";
          type = "script";
          script = ''
            echo Test something here
            # required in pure mode:
            export PATH="${lib.makeBinPath [pkgs.gnugrep]}"
            grep -q "test" ${builtins.toFile "test" "test"}
          '';
        }
      ];
    };
    "other-suite".tests = [
      {
        name = "obj-snapshot";
        type = "snapshot";
        pos = __curPos;
        actual = {hello = "world";};
      }
      {
        name = "pretty-snapshot";
        type = "snapshot";
        format = "pretty";
        pos = __curPos;
        actual = {
          example = args: {};
          example2 = {
            drv = pkgs.hello;
          };
        };
      }
      {
        name = "pretty-unit";
        format = "pretty";
        pos = __curPos;
        expected = pkgs.hello;
        actual = pkgs.hello;
      }
      {
        name = "test-drv";
        pos = __curPos;
        expected = {a = "b";};
        actualDrv = pkgs.runCommand "test-something" {} ''
          echo "Simulating taking some time"
          sleep 1
          echo '{"a":"b"}' > $out
        '';
      }
      {
        name = "skipped";
        expected = null;
        actual = null;
      }
    ];
  };
}
