{
  description = "hacompanion: Daemon that sends local hardware information to Home Assistant";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        hacompanionPkg = pkgs.buildGoModule rec {
          pname = "hacompanion";
          version = "1.0.21";
          src = ./.;

          vendorHash = "sha256-y2eSuMCDZTGdCs70zYdA8NKbuPPN5xmnRfMNK+AE/q8=";

          ldflags = [
            "-s"
            "-w"
            "-X=main.Version=${version}"
          ];

          meta = with pkgs.lib; {
            description = "Daemon that sends local hardware information to Home Assistant";
            homepage = "https://github.com/tobias-kuendig/hacompanion";
            license = licenses.mit;
            maintainers = with maintainers; [ pschmitt ];
            mainProgram = "hacompanion";
          };
        };
      in
      {
        packages.hacompanion = hacompanionPkg;
        packages.default = hacompanionPkg;

        # Enable running via `nix run github:tobias-kuendig/hacompanion`
        apps.default = {
          type = "app";
          program = "${hacompanionPkg}/bin/hacompanion";
        };
      }
    );
}
