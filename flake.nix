{
  outputs = {
    flake-parts,
    systems,
    ...
  } @ inputs:
    flake-parts.lib.mkFlake {inherit inputs;} {
      imports = [
        inputs.devenv.flakeModule
        inputs.treefmt-nix.flakeModule
        inputs.nix-gitlab-ci.flakeModule
        inputs.nix-devtools.flakeModule
        inputs.nix-mkdocs.flakeModule
        ./lib/flakeModule.nix
      ];
      systems = import systems;
      flake = {};
      perSystem = {
        lib,
        pkgs,
        config,
        ...
      }: {
        treefmt = {
          projectRootFile = "flake.nix";
          programs = {
            alejandra.enable = true;
            mdformat.enable = true;
            gofmt.enable = true;
          };
          settings.formatter.mdformat.command = let
            pkg = pkgs.python3.withPackages (p: [
              p.mdformat
              p.mdformat-mkdocs
            ]);
          in "${pkg}/bin/mdformat";
        };
        devenv.shells.default = {
          containers = pkgs.lib.mkForce {};
          packages = with pkgs; [gopls gore go-junit-report];

          languages.go.enable = true;

          pre-commit.hooks = {
            treefmt = {
              enable = true;
              packageOverrides.treefmt = config.treefmt.build.wrapper;
            };
            convco.enable = true;
          };

          task = {
            enable = true;
            alias = ",";
            tasks = {
              "test" = {
                cmds = [
                  "go test -v -coverprofile cover.out ./..."
                  "go tool cover -html cover.out -o cover.html"
                ];
              };
            };
          };
        };

        nixtest = {
          skip = "skip.*d";
          suites = {
            "suite-one" = {
              pos = __curPos;
              tests = [
                {
                  name = "test-one";
                  # required to figure out file and line, but optional
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
        };

        doc = {
          path = ./docs;
          deps = pp: [
            pp.mkdocs-material
            (pp.callPackage inputs.mkdocs-material-umami {})
          ];
          config = {
            site_name = "Nixtest";
            repo_name = "TECHNOFAB/nixtest";
            repo_url = "https://gitlab.com/TECHNOFAB/nixtest";
            edit_uri = "edit/main/docs/";
            theme = {
              name = "material";
              features = ["content.code.copy" "content.action.edit"];
              icon.repo = "simple/gitlab";
              logo = "images/logo.png";
              favicon = "images/favicon.png";
              palette = [
                {
                  scheme = "default";
                  media = "(prefers-color-scheme: light)";
                  primary = "green";
                  accent = "light green";
                  toggle = {
                    icon = "material/brightness-7";
                    name = "Switch to dark mode";
                  };
                }
                {
                  scheme = "slate";
                  media = "(prefers-color-scheme: dark)";
                  primary = "green";
                  accent = "light green";
                  toggle = {
                    icon = "material/brightness-4";
                    name = "Switch to light mode";
                  };
                }
              ];
            };
            plugins = ["search" "material-umami"];
            nav = [
              {"Introduction" = "index.md";}
              {"Usage" = "usage.md";}
              {"CLI" = "cli.md";}
              {"Example Configs" = "examples.md";}
            ];
            markdown_extensions = [
              "pymdownx.superfences"
            ];
            extra.analytics = {
              provider = "umami";
              site_id = "716d1869-9342-4b62-a770-e15d2d5c807d";
              src = "https://analytics.tf/umami";
              domains = "nixtest.projects.tf";
              feedback = {
                title = "Was this page helpful?";
                ratings = [
                  {
                    icon = "material/thumb-up-outline";
                    name = "This page is helpful";
                    data = "good";
                    note = "Thanks for your feedback!";
                  }
                  {
                    icon = "material/thumb-down-outline";
                    name = "This page could be improved";
                    data = "bad";
                    note = "Thanks for your feedback! Please leave feedback by creating an issue :)";
                  }
                ];
              };
            };
          };
        };

        ci = {
          stages = ["test" "build" "deploy"];
          jobs = {
            "test:flakeModule" = {
              stage = "test";
              script = [
                "nix run .#nixtests:run -- --junit=junit.xml"
              ];
              allow_failure = true;
              artifacts = {
                when = "always";
                reports.junit = "junit.xml";
              };
            };
            "test:lib" = {
              stage = "test";
              script = [
                "nix run .#lib-tests -- --junit=junit.xml"
              ];
              allow_failure = true;
              artifacts = {
                when = "always";
                reports.junit = "junit.xml";
              };
            };
            "test:go" = {
              stage = "test";
              nix.deps = with pkgs; [go go-junit-report gocover-cobertura];
              variables = {
                GOPATH = "$CI_PROJECT_DIR/.go";
                GOCACHE = "$CI_PROJECT_DIR/.go/pkg/mod";
              };
              script = [
                "go test -coverprofile=coverage.out -v 2>&1 ./... | go-junit-report -set-exit-code > report.xml"
                "go tool cover -func coverage.out"
                "gocover-cobertura < coverage.out > coverage.xml"
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

        packages = let
          ntlib = import ./lib {inherit pkgs lib;};
        in {
          default = pkgs.callPackage ./package.nix {};
          lib-tests = ntlib.mkNixtest {
            modules = ntlib.autodiscover {dir = ./lib;};
            args = {
              inherit pkgs;
            };
          };
        };
      };
    };

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

    # flake & devenv related
    flake-parts.url = "github:hercules-ci/flake-parts";
    systems.url = "github:nix-systems/default-linux";
    devenv.url = "github:cachix/devenv";
    treefmt-nix.url = "github:numtide/treefmt-nix";
    nix-gitlab-ci.url = "gitlab:technofab/nix-gitlab-ci/2.0.1?dir=lib";
    nix-devtools.url = "gitlab:technofab/nix-devtools?dir=lib";
    nix-mkdocs.url = "gitlab:technofab/nixmkdocs?dir=lib";
    mkdocs-material-umami.url = "gitlab:technofab/mkdocs-material-umami";
  };

  nixConfig = {
    extra-substituters = [
      "https://cache.nixos.org/"
      "https://nix-community.cachix.org"
      "https://devenv.cachix.org"
    ];

    extra-trusted-public-keys = [
      "cache.nixos.org-1:6NCHdD59X431o0gWypbMrAURkbJ16ZPMQFGspcDShjY="
      "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
      "devenv.cachix.org-1:w1cLUi8dv3hnoSPGAuibQv+f9TZLr6cv/Hm9XgU50cw="
    ];
  };
}
