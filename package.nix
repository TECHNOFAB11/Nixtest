{buildGoModule, ...}:
buildGoModule {
  name = "nixtest";
  src =
    # filter everything except for cmd/ and go.mod, go.sum
    builtins.filterSource (
      path: type:
        builtins.match ".*(/cmd/?.*|/go\.(mod|sum))$"
        path
        != null
    )
    ./.;
  subPackages = ["cmd/nixtest"];
  vendorHash = "sha256-Hmdtkp3UK/lveE2/U6FmKno38DxY+MMQlQuZFf1UBME=";
  meta.mainProgram = "nixtest";
}
