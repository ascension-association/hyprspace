## hyprspace for gokrazy

This package contains the static build of https://github.com/alecbcs/hyprspace

This is an alternative to [tailscale in gokrazy](https://gokrazy.org/packages/tailscale/). It's slower and has less features but is simpler and decentralized.

### Usage

Hyprspace is a point-to-point VPN that directly connects two machines. In the example below I'll refer to "local machine" as your home computer and "remote machine" as your gokrazy device.

1. On your local machine, initialize hyprspace:

```
cd ~
# https://github.com/alecbcs/hyprspace/releases/tag/v0.2.2
curl -fsSL -o hyprspace https://github.com/alecbcs/hyprspace/releases/download/v0.2.2/hyprspace-v0.2.2-linux-amd64
chmod +x hyprspace
touch ./hyprspace-config.yaml && chmod 600 ./hyprspace-config.yaml
./hyprspace init utun0 --config ./hyprspace-config.yaml
```

2. Get your local machine's hyprspace id: `grep "^  id:" ~/hyprspace-config.yaml`

3. Install hyprspace onto the remote machine:

```
gok add github.com/ascension-association/hyprspace
gok edit
```

4. Add the **id** from Step 2 to the _PackageConfig_ section:

```
"github.com/ascension-association/hyprspace": {
	"GoBuildFlags": [
		"-ldflags=-X main.id=QjYJafYS4zB..."
	]
}
```

5. Save and close the file, then deploy to the remote machine:

```
gok update
```

6. In the gokrazy dashboard of the remote machine, click on the link for _/user/hyprspace_ and note the **id** value in the _stdout_ section (e.g. id: QmUw6cxguRED8z...)

7. On your local machine, add the remote machine as a hyprspace peer (replacing _'QmUw6cxguRED8z...'_ with the actual gokrazy instance id):

```
sed -z 's/peers: {}/peers:\n  10.1.1.222:\n    id: QmUw6cxguRED8z.../' -i ~/hyprspace-config.yaml
```

8. Then run hyprspace:

```
# https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes
sudo sysctl -w net.core.rmem_max=2048000
sudo sysctl -w net.core.wmem_max=2048000
sudo ./hyprspace up utun0 --config ./hyprspace-config.yaml
```

9. After a moment, you should be able to ping the remote machine: `ping 10.1.1.2`

10. If successful, you can run `gok edit` and change the **Hostname** to `10.1.1.2`. Then no matter where in the world the remote machine exists as long as it has internet access you should be able to load the gokrazy web portal and/or run gok commands.

