{
  pkgs,
  lib ? pkgs.lib,
  self ? "",
  ...
}: {
  mkTest = {
    type ? "unit",
    name,
    description ? "",
    expected ? null,
    actual ? null,
    actualDrv ? null,
    pos ? null,
  }: let
    fileRelative = lib.removePrefix ((toString self) + "/") pos.file;
  in {
    inherit type name description expected actual;
    # discard string context, otherwise it's being built instantly which we don't want
    actualDrv = builtins.unsafeDiscardStringContext (actualDrv.drvPath or "");
    pos =
      if pos == null
      then ""
      else "${fileRelative}:${toString pos.line}:${toString pos.column}";
  };
  mkSuite = name: tests: {
    inherit name tests;
  };
  exportSuites = suites: let
    suitesList =
      if builtins.isList suites
      then suites
      else [suites];
    testsMapped = builtins.toJSON suitesList;
  in
    pkgs.runCommand "tests.json" {} ''
      echo '${testsMapped}' > $out
    '';
}
