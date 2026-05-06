{
  description = "modron - deterministic CLI for Notion databases";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = "0.2.0";
        src = ./.;
        vendorHash = "sha256-7K17JaXFsjf163g5PXCb5ng2gYdotnZ2IDKk8KFjNj0=";
      in
      {
        packages = {
          modron = pkgs.buildGoModule {
            pname = "modron";
            inherit version src vendorHash;
            subPackages = [ "cmd/modron" ];
            meta.description = "CLI for Notion databases";
          };

          default = self.packages.${system}.modron;
        };

        devShells.default = pkgs.mkShell {
          packages = [ pkgs.go ];
        };
      }
    );
}
