{pkgs, ...}: {
  mkTest = {
    type ? "unit",
    name,
    description ? "",
    expected ? null,
    actual ? null,
    actualDrv ? null,
    pos ? null,
  }: {
    inherit type name description expected actual;
    actualDrv = actualDrv.drvPath or "";
    pos =
      if pos == null
      then ""
      else "${pos.file}:${toString pos.line}:${toString pos.column}";
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
