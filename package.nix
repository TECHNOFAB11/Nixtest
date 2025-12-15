{
  lib,
  buildGoModule,
  ...
}:
buildGoModule {
  pname = "nixtest";
  version = "latest";
  src =
    # filter everything except for cmd/ and go.mod, go.sum
    with lib.fileset;
      toSource {
        root = ./.;
        fileset = unions [
          ./cmd
          ./internal
          ./go.mod
          ./go.sum
        ];
      };
  subPackages = ["cmd/nixtest"];
  vendorHash = "sha256-uyVSXUSoDfOhRxrtUd6KQWmx6I8kw3PJxKfYMZgz3h8=";
  meta.mainProgram = "nixtest";
}
