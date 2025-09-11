## hyprspace for gokrazy

This package contains the static build of https://github.com/alecbcs/hyprspace

This is an alternative to [tailscale in gokrazy](https://gokrazy.org/packages/tailscale/). It's slower and has less features but is simpler and decentralized.

### Usage

1. On a different machine that you want to connect to/from, initialize hyprspace:

```
cd ~
# https://github.com/alecbcs/hyprspace/releases/tag/v0.2.2
curl -fsSL -o hyprspace https://github.com/alecbcs/hyprspace/releases/download/v0.2.2/hyprspace-v0.2.2-linux-amd64
chmod +x hyprspace
touch ./hyprspace-config.yaml && chmod 600 ./hyprspace-config.yaml
./hyprspace init utun0 --config ./hyprspace-config.yaml
```

2. Get that machine's hyprspace ID: `grep "^  id:" ./hyprspace-config.yaml`

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

6. In the gokrazy dashboard, click on the link for _/user/hyprspace_ and note the **id** value in the _stdout_ section (e.g. id: QmUw6cxguRED8z...)

7. On your different machine, add the gokrazy peer:

```
sed -z 's/peers: {}/peers:\n  10.1.1.222:\n    id: QmUw6cxguRED8z.../' -i ./hyprspace-config.yaml
```

8. Then run hyprspace:

```
# https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes
sudo sysctl -w net.core.rmem_max=2048000
sudo sysctl -w net.core.wmem_max=2048000
sudo ./hyprspace up utun0 --config ./hyprspace-config.yaml
```

9. After a moment, you should be able to ping the gokrazy instance: `ping 10.1.1.222`

