{ pkgs }: {
    deps = [
        pkgs.go_1_17
		    pkgs.gotools
		    pkgs.gopls
		    pkgs.go-outline
		    pkgs.gocode
		    pkgs.gopkgs
		    pkgs.gocode-gomod
		    pkgs.godef
		    pkgs.golint
    ];
}
