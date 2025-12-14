{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    ren.url = "gitlab:rensa-nix/core?dir=lib";
  };

  outputs = {
    ren,
    self,
    ...
  } @ inputs:
    ren.buildWith
    {
      inherit inputs;
      cellsFrom = ./nix;
      transformInputs = system: i:
        i
        // {
          pkgs = import i.nixpkgs {inherit system;};
        };
      cellBlocks = with ren.blocks; [
        (simple "devShells")
        (simple "ci")
        (simple "tests")
        (simple "packages")
        (simple "docs")
        (simple "soonix")
      ];
    }
    {
      packages = ren.select self [
        ["repo" "ci" "packages"]
        ["repo" "tests"]
        ["packages" "packages"]
        ["repo" "docs"]
        ["repo" "soonix" "packages"]
      ];
    };
}
