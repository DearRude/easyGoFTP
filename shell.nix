{ pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
    in
    import (fetchTree nixpkgs.locked) {
      overlays = [ (import "${fetchTree gomod2nix.locked}/overlay.nix") ];
    }
  )
}:

let goEnv = pkgs.mkGoEnv { pwd = ./.; };
in pkgs.mkShell {
  packages = [
    pkgs.go
    pkgs.gomod2nix

    pkgs.gotools
    pkgs.gofumpt
    pkgs.golangci-lint
    pkgs.gopls
    pkgs.go-outline
    pkgs.gopkgs
  ];
  shellHook = ''
    export "GOROOT=$(go env GOROOT)"
  '';
}
