## hyprspace for gokrazy

This package contains the static build of https://github.com/alecbcs/hyprspace

This is an alternative to [tailscale in gokrazy](https://gokrazy.org/packages/tailscale/). It's slower and has less features but is simpler and decentralized.

### Usage

1. On a different machine that you want to connect to/from, run hyprspace:

```
# https://github.com/alecbcs/hyprspace/releases/tag/v0.2.2
curl -fsSL -o hyprspace https://github.com/alecbcs/hyprspace/releases/download/v0.2.2/hyprspace-v0.2.2-linux-amd64
chmod +x hyprspace
touch ./hyprspace-config.yaml && chmod 600 ./hyprspace-config.yaml
./hyprspace init utun0 --config ./hyprspace-config.yaml
```

2. Get that machine's hyprspace ID: `grep id: ./hyprspace-config.yaml`

3. On the gokrazy machine, install hyprspace:

```
gok add github.com/ascension-association/hyprspace
gok edit
```

4. Add the ID from the prior command to the _PackageConfig_ section:

```
"github.com/ascension-association/hyprspace": {
    "GoBuildFlags": [
        "-ldflags=-X main.id=QjYJafYS4zB..."
    ]
}
```

5. Deploy to the gokrazy instance:

```
gok update
```

6. In the gokrazy dashboard, click on the link for _/user/hyprspace_ and add the peer info listed in the stdout section to your different machine that you want to connect to/from

7. On your different machine, run hyprspace:

```
sudo ./hyprspace-x86_64-linux up --config ./mycomputer.json --interface hs0
```

