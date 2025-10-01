{
  pkgs,
  lib,
  ...
}: let
  inherit
    (lib)
    mkOptionType
    mkOption
    types
    filterAttrs
    isType
    removePrefix
    assertMsg
    generators
    literalExpression
    ;

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
  isUnset = isType "unset";
  unsetOr = typ:
    (types.either unsetType typ)
    // {
      inherit (typ) description getSubOptions;
    };
  mkUnsetOption = opts:
    mkOption (opts
      // {
        type = unsetOr opts.type;
        default = opts.default or unset;
        defaultText = literalExpression "unset";
      });

  filterUnset = value:
    if builtins.isAttrs value && !builtins.hasAttr "_type" value
    then let
      filteredAttrs = builtins.mapAttrs (n: v: filterUnset v) value;
    in
      filterAttrs (name: value: (!isUnset value)) filteredAttrs
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
      pos = mkUnsetOption {
        type = types.attrs;
        description = ''
          Position of test, use `__curPos` for automatic insertion of current position.
        '';
        default = pos;
        apply = val:
          if isUnset val
          then val
          else let
            fileRelative = removePrefix testsBase val.file;
          in "${fileRelative}:${toString val.line}";
      };
      type = mkOption {
        type = types.enum ["unit" "snapshot" "script"];
        description = ''
          Type of test, has to be one of "unit", "snapshot" or "script".
        '';
        default = "unit";
        apply = value:
          assert assertMsg (value != "script" || !isUnset config.script)
          "test '${config.name}' as type 'script' requires 'script' to be set";
          assert assertMsg (value != "unit" || !isUnset config.expected)
          "test '${config.name}' as type 'unit' requires 'expected' to be set";
          assert assertMsg (
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
        description = ''
          Name of this test.
        '';
      };
      description = mkUnsetOption {
        type = types.str;
        description = ''
          Short description of the test.
        '';
      };
      format = mkOption {
        type = types.enum ["json" "pretty"];
        description = ''
          Which format to use for serializing arbitrary values.
          Required since this config is serialized to JSON for passing it to Nixtest, so no Nix-values can be used directly.

          - `json`: serializes the data to json using `builtins.toJSON`
          - `pretty`: serializes the data to a "pretty" format using `lib.generators.toPretty`
        '';
        default = "json";
      };
      expected = mkUnsetOption {
        type = types.anything;
        description = ''
          Expected value of the test. Remember, the values are serialized (see [here](#suitesnametestsformat)).
        '';
        apply = val:
          if isUnset val || config.format == "json"
          then val
          else generators.toPretty {} val;
      };
      actual = mkUnsetOption {
        type = types.anything;
        description = ''
          Actual value of the test. Remember, the values are serialized (see [here](#suitesnametestsformat)).
        '';
        apply = val:
          if isUnset val || config.format == "json"
          then val
          else generators.toPretty {} val;
      };
      actualDrv = mkUnsetOption {
        type = types.package;
        description = ''
          Actual value of the test, but as a derivation.
          Nixtest will build this derivation when running the test, then compare the contents of the
          resulting file to the [`expected`](#suitesnametestsexpected) value.
        '';
        apply = val:
        # keep unset value
          if isUnset val
          then val
          else builtins.unsafeDiscardStringContext (val.drvPath or "");
      };
      script = mkUnsetOption {
        type = types.str;
        description = ''
          Script to run for the test.
          Nixtest will run this, failing the test if it exits with a non-zero exit code.
        '';
        apply = val:
          if isUnset val
          then val
          else
            builtins.unsafeDiscardStringContext
            (pkgs.writeShellScript "nixtest-${config.name}" val).drvPath;
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
        description = ''
          Name of the suite, uses attrset name by default.
        '';
        default = name;
        defaultText = literalExpression name;
      };
      pos = mkUnsetOption {
        type = types.attrs;
        description = ''
          Position for tests, use `__curPos` for automatic insertion of current position.
          This will set `pos` for every test of this suite, useful if the suite's tests are all in a single file.
        '';
        example = literalExpression "__curPos";
      };
      tests = mkOption {
        type = types.listOf (types.submoduleWith {
          modules = [testsSubmodule];
          specialArgs = {
            inherit (config) pos;
            inherit testsBase;
          };
        });
        description = ''
          Define tests of this suite here.
        '';
        default = [];
      };
    };
  };

  nixtestSubmodule = {config, ...}: {
    _file = ./module.nix;
    options = {
      base = mkOption {
        type = types.str;
        description = ''
          Base directory of the tests, will be removed from the test file path.
          This makes it possible to show the relative path from the git repo, instead of ugly Nix store paths.
        '';
        default = "";
      };
      skip = mkOption {
        type = types.str;
        description = ''
          Tests to skip, is passed to Nixtest's `--skip` param.
        '';
        default = "";
      };
      suites = mkOption {
        type = types.attrsOf (types.submoduleWith {
          modules = [suitesSubmodule];
          specialArgs = {
            testsBase = config.base;
          };
        });
        description = ''
          Define your test suites here, every test belongs to a suite.
        '';
        default = {};
        apply = suites:
          map (
            n: filterUnset (builtins.removeAttrs suites.${n} ["pos"])
          )
          (builtins.attrNames suites);
        example = {
          "Suite A".tests = [
            {
              name = "Some Test";
            }
          ];
        };
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
