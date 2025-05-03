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
  vendorHash = "sha256-H0KiuTqY2cxsUvqoxWAHKHjdfsBHjYkqxdYgTY0ftes=";
  meta.mainProgram = "nixtest";
}
