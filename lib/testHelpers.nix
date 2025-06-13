{lib, ...}: {
  path = pkgs: "export PATH=${lib.makeBinPath pkgs}";
  pathAdd = pkgs: "export PATH=$PATH:${lib.makeBinPath pkgs}";
  scriptHelpers = builtins.readFile ./scriptHelpers.sh;
}
