{
  description = "Designate cert-manager Webhook";

  inputs = {
    nixpkgs.url = "nixpkgs/6368eda62c9775c38ef7f714b2555a741c20c72d";

    kubenix = {
      url = "github:hall/kubenix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    { self, nixpkgs, ... }@inputs:
    let
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
      version = builtins.substring 0 8 lastModifiedDate;

      supportedSystems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
          name = "designate-certmanager-webhook";

          bin =
            {
              targetSystem ? system,
              version ? builtins.substring 0 8 lastModifiedDate,
            }:
            let
              pkgs =
                if targetSystem == system then
                  nixpkgsFor.${system}
                else
                  (import nixpkgs {
                    localSystem = system;
                    crossSystem = targetSystem;
                  });
            in
            pkgs.buildGoModule {
              inherit version;
              pname = name;
              src = ./.;

              env = {
                CGO_ENABLED = 0;
              };

              ldflags = [
                "-s"
                "-w"
                "-X main.gitDate=${toString (self.lastModified or 0)}"
                "-X main.gitRevision=${self.rev or self.dirtyRev or "dirty"}"
                "-X main.nixpkgsDate=${toString nixpkgs.lastModified}"
                "-X main.nixpkgsRevision=${nixpkgs.rev}"
                "-X main.version=${version}"
              ];

              vendorHash = "sha256-fOxdojgvCQOhTEWMI96FUxKaDlsib8eL9ks4ull+cUA=";

              doCheck = false; # testing depends on an external network
            };

          oci =
            let
              defaultName = name;
            in
            {
              targetSystem ? system,
              name ? defaultName,
              tag ? "latest",
            }:
            let
              pkgs =
                if targetSystem == system then
                  nixpkgsFor.${system}
                else
                  (import nixpkgs {
                    localSystem = system;
                    crossSystem = targetSystem;
                  });
            in
            pkgs.dockerTools.buildLayeredImage {
              inherit name tag;
              contents = with pkgs; [ cacert ];
              config = {
                Entrypoint = [
                  "${
                    bin {
                      inherit targetSystem;
                      version =
                        if tag == "latest" then builtins.substring 0 8 lastModifiedDate else pkgs.lib.removePrefix "v" tag;
                    }
                  }/bin/${defaultName}"
                ];
                Cmd = [ ];
              };
            };

          k8s =
            let
              app = name;
              labels = { inherit app; };
            in
            {
              values ? { },
            }:
            (inputs.kubenix.evalModules.${system} {
              module =
                { kubenix, ... }:
                {
                  imports = [ kubenix.modules.helm ];
                  kubernetes = {
                    customTypes = {
                      "certificate.cert-manager.io" = {
                        group = "cert-manager.io";
                        version = "v1";
                        attrName = "certificate";
                        kind = "Certificate";
                      };
                      "issuer.cert-manager.io" = {
                        group = "cert-manager.io";
                        version = "v1";
                        attrName = "issuer";
                        kind = "Issuer";
                      };
                    };
                    helm.releases.designate-certmanager-webhook = {
                      inherit values;
                      chart = pkgs.runCommand name {} ''
                        cp -r ${./helm/designate-certmanager-webhook}/. $out
                      '';
                    };
                  };
                };
            }).config.kubernetes.resultYAML;
        in
        {
          default = bin { };
          bin = pkgs.lib.makeOverridable bin { };
          oci = pkgs.lib.makeOverridable oci { };
          k8s = pkgs.lib.makeOverridable k8s { };
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              gotools
              go-tools
            ];
          };
        }
      );

      formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.nixfmt-tree);
    };
}
