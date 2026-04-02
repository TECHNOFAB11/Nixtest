{
  inputs,
  cell,
  ...
}: let
  inherit (inputs) pkgs devshell treefmt devtools-lib;
  inherit (cell) soonix;
  treefmtWrapper = treefmt.mkWrapper pkgs {
    projectRootFile = "flake.nix";
    programs = {
      alejandra.enable = true;
      mdformat.enable = true;
      gofmt.enable = true;
    };
    settings.formatter.mdformat = {
      excludes = ["CHANGELOG.md" "LICENSE.md"];
      command = let
        pkg = pkgs.python3.withPackages (p: [
          p.mdformat
          p.mdformat-mkdocs
        ]);
      in "${pkg}/bin/mdformat";
    };
  };
in {
  default = devshell.mkShell {
    imports = [soonix.devshellModule devtools-lib.devshellModule];
    packages = with pkgs; [
      treefmtWrapper
      gcc
      go
      gopls
      delve
      go-tools
      go-junit-report
      gocover-cobertura
    ];
    lefthook.config = {
      "pre-commit" = {
        parallel = true;
        jobs = [
          {
            name = "treefmt";
            stage_fixed = true;
            run = "${treefmtWrapper}/bin/treefmt";
            env.TERM = "dumb";
          }
          {
            name = "soonix";
            stage_fixed = true;
            run = "nix run .#soonix:update";
          }
        ];
      };
      "pre-push" = {
        parallel = true;
        jobs = [
          {
            name = "nix build";
            run = "nix build .#nixtest";
          }
          {
            name = "go test";
            run = "go test ./...";
          }
        ];
      };
    };
    cocogitto.config = {
      tag_prefix = "v";
      ignore_merge_commits = true;
      changelog = {
        authors = [
          {
            username = "TECHNOFAB";
            signature = "technofab";
          }
        ];
        path = "CHANGELOG.md";
        template = "remote";
        remote = "gitlab.com";
        repository = "nixtest";
        owner = "TECHNOFAB";
      };
    };
  };
}
