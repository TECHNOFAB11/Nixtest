{
  outputs = inputs: {
    lib = import ./.;
    flakeModule = import ./flakeModule.nix;
  };
}
