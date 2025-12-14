{inputs, ...}: let
  inherit (inputs) self pkgs;
in {
  nixtest = pkgs.callPackage "${self}/package.nix" {};
  update-package = pkgs.writeShellScriptBin "update-package" ''
    ${pkgs.nix-update}/bin/nix-update nixtest --flake --version skip
  '';
}
