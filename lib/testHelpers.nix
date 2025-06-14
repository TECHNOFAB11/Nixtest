{lib, ...}: {
  path = pkgs: "export PATH=${lib.makeBinPath pkgs}";
  pathAdd = pkgs: "export PATH=$PATH:${lib.makeBinPath pkgs}";
  scriptHelpers = builtins.readFile ./scriptHelpers.sh;
  toJsonFile = any: builtins.toFile "actual" (builtins.unsafeDiscardStringContext (builtins.toJSON any));
  toPrettyFile = any: builtins.toFile "actual" (lib.generators.toPretty {} any);
}
