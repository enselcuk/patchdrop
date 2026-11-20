# LocalShare protocol v1

## Identity

Each device generates an ECDSA P-256 self-signed certificate on first run
(stored next to `config.json` in the OS config folder). The device's verifiable
identity is the **fingerprint**: lowercase hex SHA-256 of the DER certificate.

## Discovery: UDP port 47820

Every 3 s, a JSON datagram is sent on **each** usable IPv4 interface to:

- the interface's directed broadcast address (e.g. `192.168.1.255:47820`)
- multicast group `239.255.47.82:47820` (TTL 1), sent out of that interface

```json
{"magic":"localshare","v":1,"id":"4905d93c5978ff4e","name":"Design-PC","os":"windows",
 "port":47821,"fp":"6750…","link":"wired"}
```

- `link`: kind of the sending interface (`wired`, `wifi` or empty).
- On receiving a non-`reply` announce, a device answers by unicast with
  `"reply":true`, so new devices appear on both sides at once.
- `"bye":true` is sent on shutdown, and the receiver removes the peer.
- Peers expire 10 s after their last announce.
- If a peer is heard on several addresses, a routable wired address is
  preferred over link-local (`169.254.x.x`).

Interfaces skipped: loopback, point-to-point, and virtual adapters (utun, awdl,
bridge, docker, vEthernet, VirtualBox, VMware, Tailscale, ZeroTier, WireGuard…).

## Transfer: HTTPS, TCP port 47821 (next free port up to +10)

TLS 1.3, **both sides present their device certificate** (server requires a
client cert). The client verifies the server certificate by comparing its
fingerprint with the one from discovery / the saved favorite. On first contact
by IP it is trust-on-first-use. The server learns the sender's fingerprint from
the client certificate. Auto-accept uses that fingerprint, never the
self-reported id.

| Method | Path | Body | Response |
|---|---|---|---|
| GET  | `/api/info` | | `{v,id,name,os,port,fp}` |
| POST | `/api/offer` | `{sender:{…info}, files:[{id,name,size}]}` | `200 {session, tokens:{fileId:token}}`, or `403` declined |
| PUT  | `/api/upload?session=&file=&token=` | raw file bytes, `Content-Length` = size | `204` |
| POST | `/api/cancel?session=` | | `204` |

- `name` is a slash-separated relative path (`Photos/2024/a.jpg`). The receiver
  sanitizes it: rejects `..`, strips absolute prefixes, replaces characters
  invalid on Windows, and prefixes reserved names (`CON`…).
- The offer request stays open until the user decides (2-minute timeout).
- Each token is single-use. Data is written to `<name>.part` and renamed when
  the byte count matches. On error or cancel the `.part` file is deleted.
- An existing file is never overwritten: `name (1).ext`, `name (2).ext`, … .
- Sessions idle for 10 minutes are dropped.
