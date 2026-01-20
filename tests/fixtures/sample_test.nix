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
        {
          name = "test-vm";
          type = "vm";
          vmConfig = {
            nodes.machine = {pkgs, ...}: {
              services.nginx = {
                enable = true;
                virtualHosts."localhost" = {
                  root = pkgs.writeTextDir "index.html" "Hello from nixtest VM!";
                };
              };
            };
            testScript =
              # py
              ''
                machine.wait_for_unit("nginx.service")
                machine.wait_for_open_port(80)
                machine.succeed("curl -f http://localhost | grep 'Hello from nixtest VM!'")
              '';
          };
        }
        {
          name = "vm-fail";
          type = "vm";
          vmConfig = {
            nodes.machine = {};
            testScript =
              # py
              ''
                machine.succeed("curl -f http://localhost | grep 'Hello from nixtest VM!'")
              '';
          };
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
