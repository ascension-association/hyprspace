## hyprspace for gokrazy

This package contains the static build of https://github.com/hyprspace/hyprspace

This is an alternative to [tailscale in gokrazy](https://gokrazy.org/packages/tailscale/). It's slower and has less features but is simpler and decentralized.

### Usage

1. On a different machine that you want to connect to/from, run hyprspace:

```
wget https://github.com/hyprspace/hyprspace/releases/download/v0.11.0/hyprspace-x86_64-linux
chmod +x hyprspace-x86_64-linux
sudo ./hyprspace-x86_64-linux init --config ./mycomputer.json
```

2. That command will output something like:

> Add this entry to your other peers:
> {
>   "name": "my-computer",
>   "id": "14D3KbRTkQV..."
> }

3. Install the gokrazy hyprspace client and configure:

```
gok add github.com/ascension-association/hyprspace
gok edit
```

4. Add the info from the prior command to the _PackageConfig_ section:

```
"github.com/ascension-association/hyprspace": {
    "GoBuildFlags": [
        "-ldflags=-X main.name=my-computer -X main.id=12D3KooWRFTkQV..."
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

