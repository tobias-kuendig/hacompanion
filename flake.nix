{
  description = "hacompanion: Daemon that sends local hardware information to Home Assistant";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };

        hacompanionPkg = pkgs.buildGoApplication rec {
          pname = "hacompanion";
          version = "1.0.21";
          src = ./.;
          modules = ./gomod2nix.toml;

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
