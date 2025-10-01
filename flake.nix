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
      ];
      systems = import systems;
      flake = {};
      perSystem = {
        lib,
        pkgs,
        self',
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
          packages = with pkgs; [gore go-junit-report];

          languages.go = {
            enable = true;
            enableHardeningWorkaround = true;
          };

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

        docs."default".config = {
          path = ./docs;
          material = {
            enable = true;
            colors = {
              primary = "green";
              accent = "light green";
            };
            umami = {
              enable = true;
              src = "https://analytics.tf/umami";
              siteId = "716d1869-9342-4b62-a770-e15d2d5c807d";
              domains = ["nixtest.projects.tf"];
            };
          };
          macros = {
            enable = true;
            includeDir = toString self'.packages.optionsDocs;
          };
          config = {
            site_name = "Nixtest";
            site_url = "https://nixtest.projects.tf";
            repo_name = "TECHNOFAB/nixtest";
            repo_url = "https://gitlab.com/TECHNOFAB/nixtest";
            extra_css = ["style.css"];
            theme = {
              icon.repo = "simple/gitlab";
              logo = "images/logo.svg";
              favicon = "images/logo.svg";
            };
            nav = [
              {"Introduction" = "index.md";}
              {"Usage" = "usage.md";}
              {"Reference" = "reference.md";}
              {"CLI" = "cli.md";}
              {"Example Configs" = "examples.md";}
              {"Options" = "options.md";}
            ];
            markdown_extensions = [
              "pymdownx.superfences"
              "admonition"
            ];
          };
        };

        ci = {
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
          doclib = inputs.nix-mkdocs.lib {inherit lib pkgs;};
        in rec {
          default = pkgs.callPackage ./package.nix {};
          tests = ntlib.mkNixtest {
            modules = ntlib.autodiscover {dir = ./tests;};
            args = {
              inherit pkgs ntlib;
            };
          };
          optionsDoc = doclib.mkOptionDocs {
            module = {
              _module.args.pkgs = pkgs;
              imports = [
                ntlib.module
              ];
            };
            roots = [
              {
                url = "https://gitlab.com/TECHNOFAB/nixtest/-/blob/main/lib";
                path = toString ./lib;
              }
            ];
          };
          optionsDocs = pkgs.runCommand "options-docs" {} ''
            mkdir -p $out
            ln -s ${optionsDoc} $out/options.md
          '';
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
