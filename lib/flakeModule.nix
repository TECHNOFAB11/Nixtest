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
        type = types.submodule (nixtests-lib.module);
        default = {};
      };
      config.nixtest.base = toString self + "/";

      config.legacyPackages = {
        "nixtests" = config.nixtest.finalConfigJson;
        "nixtests:run" = config.nixtest.app;
      };
    }
  );
}
