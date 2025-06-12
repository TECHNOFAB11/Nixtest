{
  pkgs,
  lib,
  ...
}: let
  inherit (lib) mkOptionType mkOption types;

  nixtest-lib = import ./default.nix {inherit pkgs lib;};

  unsetType = mkOptionType {
    name = "unset";
    description = "unset";
    descriptionClass = "noun";
    check = value: true;
  };
  unset = {
    _type = "unset";
  };
  isUnset = lib.isType "unset";

  filterUnset = value:
    if builtins.isAttrs value && !builtins.hasAttr "_type" value
    then let
      filteredAttrs = builtins.mapAttrs (n: v: filterUnset v) value;
    in
      lib.filterAttrs (name: value: (!isUnset value)) filteredAttrs
    else if builtins.isList value
    then builtins.filter (elem: !isUnset elem) (map filterUnset value)
    else value;

  testsSubmodule = {
    config,
    testsBase,
    pos,
    ...
  }: {
    options = {
      pos = mkOption {
        type = types.either types.attrs unsetType;
        default = pos;
        apply = val:
          if isUnset val
          then val
          else let
            fileRelative = lib.removePrefix testsBase val.file;
          in "${fileRelative}:${toString val.line}";
      };
      type = mkOption {
        type = types.enum ["unit" "snapshot" "script"];
        default = "unit";
        apply = value:
          assert lib.assertMsg (value != "script" || !isUnset config.script)
          "test '${config.name}' as type 'script' requires 'script' to be set";
          assert lib.assertMsg (value != "unit" || !isUnset config.expected)
          "test '${config.name}' as type 'unit' requires 'expected' to be set";
          assert lib.assertMsg (
            let
              actualIsUnset = isUnset config.actual;
              actualDrvIsUnset = isUnset config.actualDrv;
            in
              (value != "unit")
              || (!actualIsUnset && actualDrvIsUnset)
              || (actualIsUnset && !actualDrvIsUnset)
          )
          "test '${config.name}' as type 'unit' requires only 'actual' OR 'actualDrv' to be set"; value;
      };
      name = mkOption {
        type = types.str;
      };
      description = mkOption {
        type = types.either types.str unsetType;
        default = unset;
      };
      format = mkOption {
        type = types.enum ["json" "pretty"];
        default = "json";
      };
      expected = mkOption {
        type = types.anything;
        default = unset;
        apply = val:
          if isUnset val || config.format == "json"
          then val
          else lib.generators.toPretty {} val;
      };
      actual = mkOption {
        type = types.anything;
        default = unset;
        apply = val:
          if isUnset val || config.format == "json"
          then val
          else lib.generators.toPretty {} val;
      };
      actualDrv = mkOption {
        type = types.either types.package unsetType;
        default = unset;
        apply = val:
        # keep unset value
          if isUnset val
          then val
          else builtins.unsafeDiscardStringContext (val.drvPath or "");
      };
      script = mkOption {
        type = types.either types.str unsetType;
        default = unset;
        apply = val:
          if isUnset val
          then val
          else
            builtins.unsafeDiscardStringContext
            (pkgs.writeShellScript "nixtest-${config.name}" ''
              # show which line failed the test
              set -x
              ${val}
            '').drvPath;
      };
    };
  };

  suitesSubmodule = {
    name,
    config,
    testsBase,
    ...
  }: {
    options = {
      name = mkOption {
        type = types.str;
        default = name;
      };
      pos = mkOption {
        type = types.either types.attrs unsetType;
        default = unset;
      };
      tests = mkOption {
        type = types.listOf (types.submoduleWith {
          modules = [testsSubmodule];
          specialArgs = {
            inherit (config) pos;
            inherit testsBase;
          };
        });
        default = [];
      };
    };
  };

  nixtestSubmodule = {config, ...}: {
    options = {
      base = mkOption {
        description = "Base directory of the tests, will be removed from the test file path";
        type = types.str;
        default = "";
      };
      skip = mkOption {
        type = types.str;
        default = "";
      };
      suites = mkOption {
        type = types.attrsOf (types.submoduleWith {
          modules = [suitesSubmodule];
          specialArgs = {
            testsBase = config.base;
          };
        });
        default = {};
        apply = suites:
          map (
            n: filterUnset (builtins.removeAttrs suites.${n} ["pos"])
          )
          (builtins.attrNames suites);
      };

      finalConfigJson = mkOption {
        internal = true;
        type = types.package;
      };
      app = mkOption {
        internal = true;
        type = types.package;
      };
    };
    config = {
      finalConfigJson = nixtest-lib.exportSuites config.suites;
      app = nixtest-lib.mkBinary {
        nixtests = config.finalConfigJson;
        extraParams = ''--skip="${config.skip}"'';
      };
    };
  };
in
  nixtestSubmodule
