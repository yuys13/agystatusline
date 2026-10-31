{
  description = "agystatusline - Custom statusline generator for Antigravity";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{ flake-parts, treefmt-nix, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        treefmt-nix.flakeModule
      ];
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      perSystem =
        {
          config,
          self',
          inputs',
          pkgs,
          system,
          ...
        }:
        {
          treefmt = {
            projectRootFile = "flake.nix";
            settings.global.excludes = [
              ".agents/skills/gh-actions-lock/**"
            ];
            programs = {
              gofmt.enable = true;
              nixfmt.enable = true;
              prettier = {
                enable = true;
                includes = [ "*.md" ];
              };
              yamlfmt.enable = true;
            };
          };

          packages.default = pkgs.buildGoModule {
            pname = "agystatusline";
            version = "0.1.0";
            src = ./.;

            nativeCheckInputs = [ pkgs.gitMinimal ];

            vendorHash = "sha256-OE2mVO5JM2cxeROGJqctP9p7Fi3u5+P5o9tKrxddxig=";
          };

          apps.default = {
            type = "app";
            program = "${self'.packages.default}/bin/agystatusline";
            meta.description = "agystatusline binary execution app";
          };

          devShells.default = pkgs.mkShell {
            inputsFrom = [
              (self'.packages.default.overrideAttrs (_: {
                nativeCheckInputs = [ ];
              }))
            ];
            packages = with pkgs; [
              gopls
              golangci-lint
            ];
          };
        };
    };
}
