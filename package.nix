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
  vendorHash = "sha256-d3gS9VQTOFKJ0nJfT2/h2xrH93HA2730r6MO56LBEQA=";
  meta.mainProgram = "nixtest";
}
