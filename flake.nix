{
  description = "A build system for microservices";
  inputs.nixpkgs.url = "nixpkgs/nixos-22.05";


  outputs = { self, nixpkgs }:
    let
      # to work with older version of flakes
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";

      # System types to support.
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {

      # Provide some binary packages for selected system types.
      packages = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          version = "0.5.3";
        in
        {
          bob = pkgs.buildGoModule {
            pname = "bob";
            inherit version;

            # In 'nix develop', we don't need a copy of the source tree
            # in the Nix store.
            src = ./.;

            CGO_ENABLED = 0;

            ldflags = [ "-s" "-w" "-X main.Version=${version}" ];

            # This hash locks the dependencies of this package. It is
            # necessary because of how Go requires network access to resolve
            # VCS.  See https://www.tweag.io/blog/2021-03-04-gomod2nix/ for
            # details. Normally one can build with a fake sha256 and rely on native Go
            # mechanisms to tell you what the hash should be or determine what
            # it should be "out-of-band" with other tooling (eg. gomod2nix).
            # To begin with it is recommended to set this, but one must
            # remeber to bump this hash when your dependencies change.
            # vendorSha256 = pkgs.lib.fakeSha256;
            #
            # error: hash mismatch in fixed-output derivation '/nix/store/cvpva6ww5h4vy29b3zjrw9fymkfgi9kk-bob0.5.3-go-modules.drv':
            #       specified: sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
            #       got:    sha256-BzlZiAXA8wQ7RU6N1knPYH/BDX1Ae+2+4pVJ41ecK7A=*/
            #
            # If on `nix build` you get above error, just replace the value vendorSha256 with value from `got`
            vendorSha256 = "sha256-jakmXkDHjcA1BOIorrP2ZukcJhosbkJoC+Y/+wAPBCc=";

            excludedPackages = [ "example/server-db" "test/e2e" "tui-example" ];

            doCheck = false;
          };
        });

      # The default package for 'nix build'. This makes sense if the
      # flake provides only one package or there is a clear "main"
      # package.
      defaultPackage = forAllSystems (system: self.packages.${system}.bob);
    };
}
