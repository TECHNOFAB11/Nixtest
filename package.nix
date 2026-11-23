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
  vendorHash = "sha256-mF9TM3T3D6QfcbfQ/VocK13B8ZrXsd4b21xhY/hahJo=";
  meta.mainProgram = "nixtest";
}
