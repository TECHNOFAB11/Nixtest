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
    format ? "json",
    expected ? null,
    actual ? null,
    actualDrv ? null,
    script ? null,
    pos ? null,
  }: let
    fileRelative = lib.removePrefix ((toString self) + "/") pos.file;
    actual' =
      if format == "json"
      then actual
      else lib.generators.toPretty {} actual;
    expected' =
      if format == "json"
      then expected
      else lib.generators.toPretty {} expected;
  in
    assert lib.assertMsg (!(type == "script" && script == null)) "test ${name} has type 'script' but no script was passed"; {
      inherit type name description;
      actual = actual';
      expected = expected';
      # discard string context, otherwise it's being built instantly which we don't want
      actualDrv = builtins.unsafeDiscardStringContext (actualDrv.drvPath or "");
      script =
        if script != null
        then
          builtins.unsafeDiscardStringContext
          (pkgs.writeShellScript "nixtest-${name}" ''
            # show which line failed the test
            set -x
            ${script}
          '').drvPath
        else null;
      pos =
        if pos == null
        then ""
        else "${fileRelative}:${toString pos.line}";
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
