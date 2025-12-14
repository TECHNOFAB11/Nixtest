{
  inputs,
  cell,
  ...
}: let
  inherit (inputs) pkgs devshell treefmt;
  inherit (cell) soonix;
in {
  default = devshell.mkShell {
    imports = [soonix.devshellModule];
    packages = with pkgs; [
      (treefmt.mkWrapper pkgs {
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
      })
      gcc
      go
      gopls
      delve
      go-junit-report
      gocover-cobertura
    ];
  };
}
