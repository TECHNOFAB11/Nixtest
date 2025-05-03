{
  flake-parts-lib,
  lib,
  self,
  ...
}: let
  inherit (lib) mkOption types;
in {
  options.perSystem = flake-parts-lib.mkPerSystemOption (
    {
      config,
      pkgs,
      ...
    }: let
      nixtests-lib = import ./. {inherit pkgs self;};
    in {
      options.testSuites = mkOption {
        type = types.attrsOf (types.listOf types.attrs);
        default = {};
      };

      config.legacyPackages = rec {
        "nixtests" = let
          suites = map (suiteName: let
            tests = builtins.getAttr suiteName config.testSuites;
          in
            nixtests-lib.mkSuite
            suiteName
            (map (test: nixtests-lib.mkTest test) tests))
          (builtins.attrNames config.testSuites);
        in
          nixtests-lib.exportSuites suites;
        "nixtests:run" = let
          program = pkgs.callPackage ./../package.nix {};
        in
          pkgs.writeShellScriptBin "nixtests:run" ''
            ${program}/bin/nixtest --tests=${nixtests} "$@"
          '';
      };
    }
  );
}
