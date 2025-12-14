{inputs, ...}: let
  inherit (inputs) pkgs cilib;
in
  cilib.mkCI {
    pipelines."default" = {
      stages = ["test" "build" "deploy"];
      jobs = {
        "test:lib" = {
          stage = "test";
          script = [
            "nix run .#tests -- --junit=junit.xml"
          ];
          allow_failure = true;
          artifacts = {
            when = "always";
            reports.junit = "junit.xml";
          };
        };
        "test:go" = {
          stage = "test";
          nix.deps = with pkgs; [gcc go go-junit-report gocover-cobertura];
          variables = {
            GOPATH = "$CI_PROJECT_DIR/.go";
            GOCACHE = "$CI_PROJECT_DIR/.go/pkg/mod";
          };
          script = [
            # sh
            ''
              set +e
              go test -coverprofile=coverage.out -v 2>&1 ./... | go-junit-report -set-exit-code > report.xml
              TEST_EXIT_CODE=$?
              go tool cover -func coverage.out
              gocover-cobertura < coverage.out > coverage.xml

              exit $TEST_EXIT_CODE
            ''
          ];
          allow_failure = true;
          coverage = "/\(statements\)(?:\s+)?(\d+(?:\.\d+)?%)/";
          cache.paths = [".go/pkg/mod/"];
          artifacts = {
            when = "always";
            reports = {
              junit = "report.xml";
              coverage_report = {
                coverage_format = "cobertura";
                path = "coverage.xml";
              };
            };
          };
        };
        "build" = {
          stage = "build";
          script = [
            # sh
            "nix build .#nixtest"
          ];
        };
        "docs" = {
          stage = "build";
          script = [
            # sh
            ''
              nix build .#docs:default
              mkdir -p public
              cp -r result/. public/
            ''
          ];
          artifacts.paths = ["public"];
        };
        "pages" = {
          nix.enable = false;
          image = "alpine:latest";
          stage = "deploy";
          script = ["true"];
          artifacts.paths = ["public"];
          rules = [
            {
              "if" = "$CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH";
            }
          ];
        };
      };
    };
  }
