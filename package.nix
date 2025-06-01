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
          ./go.mod
          ./go.sum
        ];
      };
  subPackages = ["cmd/nixtest"];
  vendorHash = "sha256-Hmdtkp3UK/lveE2/U6FmKno38DxY+MMQlQuZFf1UBME=";
  meta.mainProgram = "nixtest";
}
