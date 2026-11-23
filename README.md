# PatchDrop

Fast file and folder sharing between computers on the same local network,
built for **wired (Ethernet) office networks** where LocalSend-style discovery
often fails. Works on macOS and Windows. No internet, no accounts, no cloud.

Website and downloads: <https://patchdrop.enginsistem.com>

**Beta: tested in our office on Mac and Windows.**

## Why it works on wired networks

LocalSend announces itself with a single multicast packet that the OS sends out
of one "default" interface (often Wi-Fi, a VPN or a virtual adapter), and many
switches drop multicast (IGMP snooping). PatchDrop instead:

1. Sends a **directed broadcast** (e.g. `192.168.1.255`) on **every real
   interface**: Ethernet first, skipping VPN/Docker/VM adapters. Every switch
   forwards broadcast within the subnet.
2. Also sends multicast on each interface as a second channel.
3. Has **Add by IP** (saved as a favorite) and **Scan network** (probes the
   subnet directly) for networks where broadcast is blocked too.

Transfers are encrypted (TLS 1.3). Each device has its own certificate, and
PatchDrop refuses to send when a device's certificate doesn't match the one it
expects. See [PROTOCOL.md](PROTOCOL.md).

## Install

Download from <https://patchdrop.enginsistem.com> or the
[latest release](https://github.com/enselcuk/patchdrop/releases/latest). Requires macOS 12 or later, or Windows 10/11.

| OS      | File                    | Notes |
|---------|-------------------------|-------|
| macOS   | `PatchDrop-macOS.dmg` (Intel + Apple Silicon) | Copy to /Applications. Signed and notarized by Apple. Allow **Local Network** access when asked. |
| Windows | `PatchDrop-Windows-Setup-x64.exe` (installer) | Installs to Program Files with Start menu and desktop shortcuts, adds the firewall rule for Private/Domain networks, and installs WebView2 if missing. Uninstall from Settings → Apps. |
| Windows | `PatchDrop-Windows-x64.zip` (portable) | Needs the WebView2 runtime (preinstalled on Windows 10/11). When the firewall asks, allow **Private networks**. |

### Firewall (if devices don't see each other)

PatchDrop uses **UDP 47820** (discovery) and **TCP 47821** (transfers).

Windows, in an admin PowerShell:

```powershell
New-NetFirewallRule -DisplayName "PatchDrop" -Direction Inbound -Program "C:\Path\To\PatchDrop.exe" -Action Allow -Profile Private,Domain
```

Also make sure the Ethernet network is set to **Private**, not Public
(Settings → Network → Ethernet → Network profile type).

macOS: System Settings → Network → Firewall → Options → allow PatchDrop.

## Use

- Devices running PatchDrop appear on the left. The plug icon means it was
  reached over a wired link.
- Drop files/folders on the window (or **Choose files / Choose folder**),
  select a device, click **Send**.
- The receiver gets an Accept / Decline prompt. Tick **Always accept from this
  device** for trusted coworkers.
- Received files go to `~/Downloads/PatchDrop` (change it in Settings). Name
  clashes become `file (1).ext`, so nothing is overwritten.
- ★ a device to remember it even when discovery is unavailable.

## Command line (macOS)

The same binary has a small CLI, handy for scripts and troubleshooting:

```bash
APP=/Applications/PatchDrop.app/Contents/MacOS/PatchDrop
$APP interfaces                         # which interfaces are used, wired/wifi
$APP peers                              # list devices found in 8 s
$APP send 192.168.1.24 ./report.pdf ./Photos
$APP receive                            # headless receiver, accepts everything
```

(On Windows the .exe is a GUI program, so CLI output is not shown in a console.)

## Develop

Requirements: Go 1.26+, Node 20+, Wails v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).

```bash
wails dev                                   # live-reload dev mode
go test ./internal/...                      # unit + loopback integration tests
wails build -platform darwin/universal      # macOS app
wails build -platform windows/amd64 -o PatchDrop.exe   # Windows (cross-compiles from a Mac)
```

Windows installer (NSIS). Run `wails build -platform windows/amd64 -nsis -o PatchDrop.exe`
once to fill in `wails_tools.nsh` and download the WebView2 bootstrapper. Homebrew's
`makensis` 3.13 crashes on macOS 27, so compile the installer with Linux NSIS in Docker:

```bash
docker run --rm -v "$PWD/build:/build" -w /build/windows/installer debian:stable-slim \
  sh -c 'apt-get update -qq && apt-get install -y -qq nsis && makensis -DARG_WAILS_AMD64_BINARY=/build/bin/PatchDrop.exe project.nsi'
# → build/bin/PatchDrop-Windows-Setup-x64.exe
```

Run two instances on one machine for testing:

```bash
PATCHDROP_PROFILE=a wails dev
PATCHDROP_PROFILE=b PATCHDROP_PORT=47831 ./build/bin/PatchDrop.app/Contents/MacOS/PatchDrop receive
```

### Layout

```
main.go, app.go, cli.go   Wails entry, UI bindings, CLI
internal/netif            interface enumeration, wired/wifi detection, broadcast addresses
internal/discovery        UDP announce/listen, peer table
internal/identity         device certificate + fingerprint
internal/server           HTTPS receiver (offer / upload / cancel)
internal/client           sender (pinned TLS, streaming upload)
internal/node             glues it all together; favorites, scan, transfers
frontend/src              Svelte 5 UI
```

## License

MIT, © 2026 Engin Tasarım Ltd. Şti. See [LICENSE](LICENSE).
