{
  lib,
  buildGoModule,
  ...
}:
buildGoModule {
  name = "nixtest";
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
  vendorHash = "sha256-6kARJgngmXielUoXukYdAA0QHk1mwLRvgKJhx+v1iSo=";
  meta.mainProgram = "nixtest";
}
