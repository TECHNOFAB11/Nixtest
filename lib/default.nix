{
  pkgs,
  lib ? pkgs.lib,
  ...
}: let
  inherit (lib) evalModules toList;
in rec {
  helpers = import ./testHelpers.nix {inherit lib;};

  mkBinary = {
    nixtests,
    extraParams,
  }: let
    program = pkgs.callPackage ../package.nix {};
  in
    (pkgs.writeShellScriptBin "nixtests:run" ''
      ${program}/bin/nixtest --tests=${nixtests} ${extraParams} "$@"
    '')
    // {
      tests = nixtests;
    };

  exportSuites = suites: let
    suitesList =
      if builtins.isList suites
      then suites
      else [suites];
    suitesMapped = builtins.toJSON suitesList;
  in
    pkgs.runCommand "tests.json" {} ''
      echo '${suitesMapped}' > $out
    '';

  module = import ./module.nix {inherit lib pkgs;};

  autodiscover = {
    dir,
    pattern ? ".*_test.nix",
  }: let
    files = builtins.readDir dir;
    matchingFiles = builtins.filter (name: builtins.match pattern name != null) (builtins.attrNames files);
    imports = map (file:
      if builtins.isString dir
      then (builtins.unsafeDiscardStringContext dir) + "/${file}"
      else /${dir}/${file})
    matchingFiles;
  in {
    inherit imports;
    # automatically set the base so test filepaths are easier to read
    config.base = builtins.toString dir + "/";
  };

  mkNixtestConfig = {
    modules,
    args ? {},
    ...
  }:
    (evalModules {
      modules =
        (toList modules)
        ++ [
          module
          {
            _module.args = args;
          }
        ];
    }).config;

  mkNixtest = args: (mkNixtestConfig args).app;
}
