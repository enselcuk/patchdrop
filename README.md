# LocalShare

Fast file and folder sharing between computers on the same local network,
built for **wired (Ethernet) office networks** where LocalSend-style discovery
often fails. Works on macOS and Windows. No internet, no accounts, no cloud.

## Why it works on wired networks

LocalSend announces itself with a single multicast packet that the OS sends out
of one "default" interface (often Wi-Fi, a VPN or a virtual adapter), and many
switches drop multicast (IGMP snooping). LocalShare instead:

1. Sends a **directed broadcast** (e.g. `192.168.1.255`) on **every real
   interface**: Ethernet first, skipping VPN/Docker/VM adapters. Every switch
   forwards broadcast within the subnet.
2. Also sends multicast on each interface as a second channel.
3. Has **Add by IP** (saved as a favorite) and **Scan network** (probes the
   subnet directly) for networks where broadcast is blocked too.

Transfers are HTTPS with **mutual TLS**. Each device has a self-signed
certificate, and its fingerprint is pinned, so another machine on the LAN cannot
impersonate a coworker. See [PROTOCOL.md](PROTOCOL.md).

## Install

Grab the builds from `build/bin/`:

| OS      | File                    | Notes |
|---------|-------------------------|-------|
| macOS   | `localshare.app` (Intel + Apple Silicon) | Copy to /Applications. First launch: right-click → **Open** (the app is not notarized). Allow **Local Network** access when asked. |
| Windows | `LocalShare.exe`        | Needs the WebView2 runtime (preinstalled on Windows 10/11). When the firewall asks, allow **Private networks**. |

### Firewall (if devices don't see each other)

LocalShare uses **UDP 47820** (discovery) and **TCP 47821** (transfers).

Windows, in an admin PowerShell:

```powershell
New-NetFirewallRule -DisplayName "LocalShare" -Direction Inbound -Program "C:\Path\To\LocalShare.exe" -Action Allow -Profile Private,Domain
```

Also make sure the Ethernet network is set to **Private**, not Public
(Settings → Network → Ethernet → Network profile type).

macOS: System Settings → Network → Firewall → Options → allow LocalShare.

## Use

- Devices running LocalShare appear on the left. The plug icon means it was
  reached over a wired link.
- Drop files/folders on the window (or **Choose files / Choose folder**),
  select a device, click **Send**.
- The receiver gets an Accept / Decline prompt. Tick **Always accept from this
  device** for trusted coworkers.
- Received files go to `~/Downloads/LocalShare` (change it in Settings). Name
  clashes become `file (1).ext`, so nothing is overwritten.
- ★ a device to remember it even when discovery is unavailable.

## Command line (macOS)

The same binary has a small CLI, handy for scripts and troubleshooting:

```bash
APP=/Applications/localshare.app/Contents/MacOS/localshare
$APP interfaces                         # which interfaces are used, wired/wifi
$APP peers                              # list devices found in 8 s
$APP send 192.168.1.24 ./report.pdf ./Photos
$APP receive                            # headless receiver, accepts everything
```

(On Windows the .exe is a GUI program, so CLI output is not shown in a console.)

## Develop

Requirements: Go 1.24+, Node 20+, Wails v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).

```bash
wails dev                                   # live-reload dev mode
go test ./internal/...                      # unit + loopback integration tests
wails build -platform darwin/universal      # macOS app
wails build -platform windows/amd64 -o LocalShare.exe   # Windows (cross-compiles from a Mac)
```

Run two instances on one machine for testing:

```bash
LOCALSHARE_PROFILE=a wails dev
LOCALSHARE_PROFILE=b LOCALSHARE_PORT=47831 ./build/bin/localshare.app/Contents/MacOS/localshare receive
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
