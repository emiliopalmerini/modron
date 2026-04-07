{
  description = "Notion MCP server - deterministic tools for Notion databases";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = "0.1.0";
        src = ./.;
        vendorHash = "sha256-/m9PZgfJsUfSqQs5e4HP/KpGVtfDiK/v20G+vTI3vS4=";
      in
      {
        packages = {
          notion-mcp = pkgs.buildGoModule {
            pname = "notion-mcp";
            inherit version src vendorHash;
            subPackages = [ "cmd/notion-mcp" ];
            meta.description = "MCP server for Notion databases";
          };

          default = self.packages.${system}.notion-mcp;
        };

        devShells.default = pkgs.mkShell {
          packages = [ pkgs.go ];
        };
      }
    );
}
