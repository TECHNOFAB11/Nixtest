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
        ./lib/flakeModule.nix
      ];
      systems = import systems;
      flake = {};
      perSystem = {
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
        };
        devenv.shells.default = {
          containers = pkgs.lib.mkForce {};
          packages = [pkgs.gopls pkgs.gore];

          languages.go.enable = true;

          pre-commit.hooks = {
            treefmt = {
              enable = true;
              packageOverrides.treefmt = config.treefmt.build.wrapper;
            };
            convco.enable = true;
          };
        };

        testSuites = {
          "suite-one" = [
            {
              name = "test-one";
              # required to figure out file and line, but optional
              pos = __curPos;
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
              pos = __curPos;
              actual = "test";
            }
          ];
          "other-suite" = [
            {
              name = "obj-snapshot";
              type = "snapshot";
              actual = {hello = "world";};
            }
          ];
        };

        ci = {
          stages = ["test"];
          jobs = {
            "test" = {
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
          };
        };

        packages.default = pkgs.callPackage ./package.nix {};
      };
    };

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

    # flake & devenv related
    flake-parts.url = "github:hercules-ci/flake-parts";
    systems.url = "github:nix-systems/default-linux";
    devenv.url = "github:cachix/devenv";
    treefmt-nix.url = "github:numtide/treefmt-nix";
    nix-gitlab-ci.url = "gitlab:TECHNOFAB/nix-gitlab-ci/2.0.0?dir=lib";
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
