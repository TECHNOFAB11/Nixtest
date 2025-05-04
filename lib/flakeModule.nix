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
      options.nixtest = mkOption {
        type = types.submodule ({...}: {
          options = {
            skip = mkOption {
              type = types.str;
              default = "";
              description = "Which tests to skip (regex)";
            };
            suites = mkOption {
              type = types.attrsOf (types.listOf types.attrs);
              default = {};
            };
          };
        });
        default = {};
      };

      config.legacyPackages = rec {
        "nixtests" = let
          suites = map (suiteName: let
            tests = builtins.getAttr suiteName config.nixtest.suites;
          in
            nixtests-lib.mkSuite
            suiteName
            (map (test: nixtests-lib.mkTest test) tests))
          (builtins.attrNames config.nixtest.suites);
        in
          nixtests-lib.exportSuites suites;
        "nixtests:run" = let
          program = pkgs.callPackage ./../package.nix {};
        in
          pkgs.writeShellScriptBin "nixtests:run" ''
            ${program}/bin/nixtest --tests=${nixtests} --skip="${config.nixtest.skip}" "$@"
          '';
      };
    }
  );
}
