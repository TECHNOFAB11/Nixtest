{
  pkgs,
  lib,
  ...
}: let
  ntlib = import ./. {inherit pkgs lib;};
in {
  suites."Lib Tests".tests = [
    {
      name = "autodiscovery";
      type = "script";
      script = let
        actual = builtins.toFile "actual" (builtins.toJSON (ntlib.autodiscover {
          dir = ./.;
        }));
      in
        # sh
        ''
          export PATH="${pkgs.gnugrep}/bin"
          grep -q lib_test.nix ${actual}
          grep -q "\"base\":\"/nix/store/.*-source/lib/" ${actual}
        '';
    }
    {
      name = "binary";
      type = "script";
      script = let
        binary =
          (ntlib.mkBinary {
            nixtests = "stub";
            extraParams = "--pure";
          })
          + "/bin/nixtests:run";
      in
        # sh
        ''
          export PATH="${pkgs.gnugrep}/bin"
          grep -q nixtest ${binary}
          grep -q -- "--pure" ${binary}
          grep -q -- "--tests=stub" ${binary}
        '';
    }
  ];
}
