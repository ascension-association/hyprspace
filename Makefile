all: _gokrazy/extrafiles_arm64.tar _gokrazy/extrafiles_amd64.tar

_gokrazy/extrafiles_amd64.tar:
	mkdir -p _gokrazy/extrafiles_amd64/usr/local/bin
	curl -fsSL -o _gokrazy/extrafiles_amd64/usr/local/bin/hyprspace https://github.com/alecbcs/hyprspace/releases/download/v0.2.2/hyprspace-v0.2.2-linux-amd64
	chmod +x _gokrazy/extrafiles_amd64/usr/local/bin/hyprspace
	cd _gokrazy/extrafiles_amd64 && tar cf ../extrafiles_amd64.tar *
	rm -rf _gokrazy/extrafiles_amd64

_gokrazy/extrafiles_arm64.tar:
	mkdir -p _gokrazy/extrafiles_arm64/usr/local/bin
	curl -fsSL -o _gokrazy/extrafiles_arm64/usr/local/bin/hyprspace https://github.com/alecbcs/hyprspace/releases/download/v0.2.2/hyprspace-v0.2.2-linux-arm64
	chmod +x _gokrazy/extrafiles_arm64/usr/local/bin/hyprspace
	cd _gokrazy/extrafiles_arm64 && tar cf ../extrafiles_arm64.tar *
	rm -rf _gokrazy/extrafiles_arm64

clean:
	rm -f _gokrazy/extrafiles_*.tar
