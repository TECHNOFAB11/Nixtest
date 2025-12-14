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
  vendorHash = "sha256-vH9lQqLWeaNIvWEZs7uPmPL6cINBzrteOJaIMgdRXZM=";
  meta.mainProgram = "nixtest";
}
