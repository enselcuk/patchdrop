<script lang="ts">
  import { onMount } from 'svelte'
  import * as api from '../wailsjs/go/main/App'
  import { EventsOn, OnFileDrop } from '../wailsjs/runtime/runtime'
  import Icon from './lib/Icon.svelte'
  import PeerList from './lib/PeerList.svelte'
  import TransferList from './lib/TransferList.svelte'
  import Modal from './lib/Modal.svelte'
  import type { Peer, Transfer, Self, Offer } from './types'
  import { bytes, baseName, osLabel } from './format'

  let self = $state<Self | null>(null)
  let fatal = $state('')
  let peers = $state<Peer[]>([])
  let transfers = $state<Record<string, Transfer>>({})
  let offers = $state<Offer[]>([])
  let selected = $state('')
  let staged = $state<string[]>([])
  let addr = $state('')
  let adding = $state(false)
  let scanning = $state(false)
  let showSettings = $state(false)
  let nameDraft = $state('')
  let alwaysAccept = $state(false)
  let toasts = $state<{ id: number; text: string; kind: 'error' | 'info' }[]>([])

  const selectedPeer = $derived(peers.find((p) => p.id === selected))
  const transferList = $derived(Object.values(transfers).sort((a, b) => b.startedAt - a.startedAt))
  const offer = $derived(offers[0])

  let toastSeq = 0
  function toast(text: string, kind: 'error' | 'info' = 'error') {
    const id = ++toastSeq
    toasts.push({ id, text, kind })
    setTimeout(() => (toasts = toasts.filter((t) => t.id !== id)), kind === 'error' ? 6000 : 3500)
  }
  const errText = (e: unknown) => (typeof e === 'string' ? e : (e as Error)?.message ?? String(e))

  onMount(() => {
    api.GetState().then((s) => {
      if (s.error) {
        fatal = s.error
        return
      }
      self = s.self as Self
      peers = (s.peers ?? []) as Peer[]
      for (const t of (s.transfers ?? []) as Transfer[]) transfers[t.id] = t
      if (!selected && peers.length) selected = peers.find((p) => p.online)?.id ?? ''
    })
    const offs = [
      EventsOn('peers', (p: Peer[]) => {
        peers = p ?? []
        if (!selected) selected = peers.find((x) => x.online)?.id ?? ''
      }),
      EventsOn('transfer', (t: Transfer) => {
        const prev = transfers[t.id]
        transfers[t.id] = t
        if (prev?.status !== t.status && t.status === 'done' && t.direction === 'receive') {
          toast(`Received ${t.label} from ${t.peerName}`, 'info')
        }
      }),
      EventsOn('offer', (o: Offer) => {
        offers.push(o)
      }),
      EventsOn('offerClosed', (id: string) => {
        offers = offers.filter((o) => o.id !== id)
      }),
    ]
    OnFileDrop((_x, _y, paths) => stage(paths), true)
    return () => offs.forEach((off) => off())
  })

  function stage(paths: string[]) {
    for (const p of paths) if (!staged.includes(p)) staged.push(p)
  }

  async function pickFiles() {
    try { stage((await api.PickFiles()) ?? []) } catch (e) { toast(errText(e)) }
  }
  async function pickFolder() {
    try {
      const d = await api.PickFolder()
      if (d) stage([d])
    } catch (e) { toast(errText(e)) }
  }

  async function send() {
    if (!selectedPeer || staged.length === 0) return
    try {
      await api.Send(selectedPeer.id, $state.snapshot(staged))
      staged = []
    } catch (e) {
      toast(errText(e))
    }
  }

  async function addByIP(e: Event) {
    e.preventDefault()
    if (!addr.trim()) return
    adding = true
    try {
      const p = (await api.AddByAddress(addr)) as Peer
      selected = p.id
      addr = ''
      toast(`Added ${p.name}`, 'info')
    } catch (e) {
      toast(errText(e))
    } finally {
      adding = false
    }
  }

  async function scan() {
    scanning = true
    try {
      const n = await api.Scan()
      toast(n === 0 ? 'No other LocalShare devices answered on this network.' : `Found ${n} device${n === 1 ? '' : 's'}`, n === 0 ? 'error' : 'info')
    } catch (e) {
      toast(errText(e))
    } finally {
      scanning = false
    }
  }

  async function toggleFavorite(p: Peer) {
    try { await api.SetFavorite(p.id, !p.favorite, p.autoAccept) } catch (e) { toast(errText(e)) }
  }

  async function respond(accept: boolean) {
    const o = offer
    if (!o) return
    offers = offers.slice(1)
    api.RespondOffer(o.id, accept)
    if (accept && alwaysAccept) {
      api.SetFavorite(o.sender.id, true, true).catch(() => {})
    }
    alwaysAccept = false
  }

  function openSettings() {
    nameDraft = self?.name ?? ''
    showSettings = true
  }
  async function saveName() {
    try {
      await api.SetDeviceName(nameDraft)
      if (self) self.name = nameDraft.trim()
    } catch (e) { toast(errText(e)) }
  }
  async function changeDir() {
    try {
      const d = await api.ChooseDownloadDir()
      if (d && self) self.downloadDir = d
    } catch (e) { toast(errText(e)) }
  }
  const reveal = (p: string) => api.Reveal(p).catch((e) => toast(errText(e)))
</script>

{#if fatal}
  <div class="fatal">
    <h2>LocalShare could not start</h2>
    <p class="mono">{fatal}</p>
    <p class="muted">Another copy may already be running, or the network ports are blocked.</p>
  </div>
{:else}
  <div class="shell">
    <header>
      <div class="brand"><span class="logo"><Icon name="upload" size={16} /></span> LocalShare</div>
      {#if self}
        <button class="me ghost" onclick={openSettings} title="Settings">
          <span class="dot"></span>
          <span><strong>{self.name}</strong> <span class="muted">· {self.ips[0] ?? 'no network'}</span></span>
          <Icon name="gear" size={15} />
        </button>
      {/if}
    </header>

    <aside>
      <div class="side-head">
        <h2>Devices</h2>
        <button class="ghost" title="Refresh" aria-label="Refresh" onclick={() => api.Refresh()}><Icon name="refresh" size={15} /></button>
      </div>
      <div class="peers">
        <PeerList {peers} bind:selected onfavorite={toggleFavorite} />
      </div>
      <div class="manual">
        <form onsubmit={addByIP}>
          <input type="text" placeholder="Add by IP, e.g. 192.168.1.24" bind:value={addr} disabled={adding} spellcheck="false" />
          <button type="submit" disabled={adding || !addr.trim()} title="Add device" aria-label="Add device"><Icon name="plus" size={15} /></button>
        </form>
        <button class="scan" onclick={scan} disabled={scanning}>
          <span class:spin={scanning}><Icon name="radar" size={15} /></span>
          {scanning ? 'Scanning network…' : 'Scan network'}
        </button>
      </div>
    </aside>

    <main>
      <section class="drop" style="--wails-drop-target: drop">
        {#if staged.length === 0}
          <div class="drop-empty">
            <div class="drop-icon"><Icon name="upload" size={28} /></div>
            <p class="big">Drop files or folders here</p>
            <div class="pick">
              <button onclick={pickFiles}><Icon name="file" size={14} /> Choose files</button>
              <button onclick={pickFolder}><Icon name="folder" size={14} /> Choose folder</button>
            </div>
          </div>
        {:else}
          <div class="staged">
            <div class="staged-head">
              <strong>{staged.length} item{staged.length === 1 ? '' : 's'} ready</strong>
              <span>
                <button class="ghost" onclick={pickFiles}>+ Files</button>
                <button class="ghost" onclick={pickFolder}>+ Folder</button>
                <button class="ghost" onclick={() => (staged = [])}>Clear</button>
              </span>
            </div>
            <ul>
              {#each staged as p (p)}
                <li>
                  <span class="fname" title={p}>{baseName(p)}</span>
                  <button class="ghost" aria-label="Remove" onclick={() => (staged = staged.filter((x) => x !== p))}><Icon name="x" size={13} /></button>
                </li>
              {/each}
            </ul>
            <div class="send-row">
              {#if selectedPeer}
                <span class="muted">to <strong class="to">{selectedPeer.name}</strong>{selectedPeer.online ? '' : ' (offline)'}</span>
              {:else}
                <span class="muted">Select a device on the left</span>
              {/if}
              <button class="primary" disabled={!selectedPeer?.online} onclick={send}>Send</button>
            </div>
          </div>
        {/if}
      </section>

      <section class="transfers">
        <TransferList
          transfers={transferList}
          oncancel={(id) => api.CancelTransfer(id)}
          onreveal={reveal}
          onclear={() => {
            api.ClearFinished()
            for (const t of transferList) if (!['waiting', 'active'].includes(t.status)) delete transfers[t.id]
          }}
        />
      </section>
    </main>
  </div>

  {#if offer}
    <Modal title="Incoming files">
      <p class="offer-lead">
        <strong>{offer.sender.name}</strong>
        <span class="muted">({osLabel(offer.sender.os)} · {offer.senderIp})</span>
        wants to send you {offer.files.length} file{offer.files.length === 1 ? '' : 's'} · <strong>{bytes(offer.totalBytes)}</strong>
      </p>
      <ul class="offer-files">
        {#each offer.files.slice(0, 8) as f (f.id)}
          <li><span class="fname">{f.name}</span><span class="muted">{bytes(f.size)}</span></li>
        {/each}
        {#if offer.files.length > 8}<li class="muted">+ {offer.files.length - 8} more</li>{/if}
      </ul>
      <label class="check"><input type="checkbox" bind:checked={alwaysAccept} /> Always accept from this device</label>
      <p class="muted fp mono" title="The sender's certificate fingerprint">ID {offer.sender.fp.slice(0, 16).replace(/(.{4})/g, '$1 ')}</p>
      <div class="modal-actions">
        <button onclick={() => respond(false)}>Decline</button>
        <button class="primary" onclick={() => respond(true)}>Accept</button>
      </div>
      {#if offers.length > 1}<p class="muted more">{offers.length - 1} more waiting</p>{/if}
    </Modal>
  {/if}

  {#if showSettings && self}
    <Modal title="Settings" onclose={() => (showSettings = false)}>
      <label class="field">
        <span>Device name</span>
        <div class="inline">
          <input type="text" bind:value={nameDraft} maxlength="64" />
          <button onclick={saveName} disabled={!nameDraft.trim() || nameDraft.trim() === self.name}>Save</button>
        </div>
      </label>
      <div class="field">
        <span>Save received files to</span>
        <div class="inline">
          <div class="path mono" title={self.downloadDir}>{self.downloadDir}</div>
          <button onclick={changeDir}>Change</button>
          <button onclick={() => reveal('')} title="Open folder" aria-label="Open folder"><Icon name="open" size={14} /></button>
        </div>
      </div>
      <div class="field">
        <span>This device</span>
        <div class="kv mono">
          {#each self.ips as ip}<div>{ip}:{self.port}</div>{/each}
          {#if self.ips.length === 0}<div>No network connection</div>{/if}
        </div>
        <p class="muted hint">
          Others can add this computer with "Add by IP" using one of these addresses.
          LocalShare uses UDP {47820} (discovery) and TCP {self.port} (transfers). Allow both in your firewall.
        </p>
      </div>
      <div class="field">
        <span>Device ID (certificate fingerprint)</span>
        <div class="mono fpfull">{self.fp.replace(/(.{4})/g, '$1 ')}</div>
      </div>
      <div class="modal-actions">
        <button class="primary" onclick={() => (showSettings = false)}>Done</button>
      </div>
    </Modal>
  {/if}

  <div class="toasts" aria-live="polite">
    {#each toasts as t (t.id)}
      <div class="toast {t.kind}">{t.text}</div>
    {/each}
  </div>
{/if}

<style>
  .shell {
    display: grid;
    grid-template-columns: 290px 1fr;
    grid-template-rows: 52px 1fr;
    height: 100vh;
  }
  header {
    grid-column: 1 / -1;
    display: flex; align-items: center; justify-content: space-between;
    padding: 0 14px 0 18px;
    border-bottom: 1px solid var(--border);
    background: var(--panel);
    --wails-draggable: drag;
  }
  .brand { display: flex; align-items: center; gap: 9px; font-weight: 700; font-size: 15px; }
  .logo { width: 26px; height: 26px; border-radius: 7px; background: var(--accent); color: var(--accent-text); display: grid; place-items: center; }
  .me { display: flex; align-items: center; gap: 8px; --wails-draggable: no-drag; font-size: 13px; }
  .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--ok); }

  aside {
    border-right: 1px solid var(--border);
    background: var(--panel);
    display: flex; flex-direction: column; min-height: 0;
  }
  .side-head { display: flex; align-items: center; justify-content: space-between; padding: 14px 12px 6px 16px; }
  h2 { font-size: 13px; text-transform: uppercase; letter-spacing: 0.05em; color: var(--muted); margin: 0; }
  .peers { flex: 1; overflow-y: auto; padding: 4px 8px; }
  .manual { border-top: 1px solid var(--border); padding: 12px; display: flex; flex-direction: column; gap: 8px; }
  .manual form { display: flex; gap: 6px; }
  .manual form button { padding: 7px 10px; display: flex; align-items: center; }
  .scan { display: flex; align-items: center; justify-content: center; gap: 7px; }
  .spin { display: flex; animation: spin 1s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }

  main { display: flex; flex-direction: column; min-height: 0; overflow-y: auto; padding: 18px; gap: 18px; }
  .drop {
    border: 2px dashed var(--border); border-radius: 14px; background: var(--panel);
    min-height: 220px; display: flex; transition: border-color 0.15s, background 0.15s;
  }
  :global(.wails-drop-target-active).drop { border-color: var(--accent); background: var(--accent-soft); }
  .drop-empty { margin: auto; text-align: center; padding: 24px; }
  .drop-icon { width: 56px; height: 56px; border-radius: 14px; margin: 0 auto 10px; display: grid; place-items: center; background: var(--accent-soft); color: var(--accent); }
  .big { font-size: 16px; font-weight: 600; margin: 4px 0 14px; }
  .pick { display: flex; gap: 8px; justify-content: center; }
  .pick button { display: flex; align-items: center; gap: 6px; }

  .staged { flex: 1; display: flex; flex-direction: column; padding: 14px; gap: 10px; min-width: 0; }
  .staged-head { display: flex; justify-content: space-between; align-items: center; }
  .staged ul { list-style: none; margin: 0; padding: 0; max-height: 170px; overflow-y: auto; display: flex; flex-direction: column; gap: 2px; }
  .staged li { display: flex; align-items: center; justify-content: space-between; padding: 4px 4px 4px 10px; border-radius: 6px; background: var(--panel-2); }
  .fname { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; min-width: 0; }
  .send-row { display: flex; align-items: center; justify-content: space-between; gap: 10px; margin-top: auto; padding-top: 6px; }
  .send-row .primary { min-width: 110px; }
  .to { color: var(--text); }

  .offer-lead { margin: 0 0 12px; line-height: 1.5; }
  .offer-files { list-style: none; padding: 0; margin: 0 0 12px; border: 1px solid var(--border); border-radius: 8px; max-height: 220px; overflow-y: auto; }
  .offer-files li { display: flex; justify-content: space-between; gap: 12px; padding: 6px 10px; font-size: 13px; border-bottom: 1px solid var(--border); }
  .offer-files li:last-child { border-bottom: 0; }
  .check { display: flex; gap: 8px; align-items: center; font-size: 13px; }
  .fp { font-size: 11px; margin: 8px 0 0; }
  .more { font-size: 12px; text-align: right; margin: 8px 0 0; }
  .modal-actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 16px; }

  .field { display: flex; flex-direction: column; gap: 6px; margin-bottom: 14px; }
  .field > span { font-size: 12px; font-weight: 600; color: var(--muted); }
  .inline { display: flex; gap: 6px; align-items: center; }
  .path { flex: 1; min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; padding: 8px 10px; background: var(--panel-2); border: 1px solid var(--border); border-radius: 8px; user-select: text; -webkit-user-select: text; }
  .kv { user-select: text; -webkit-user-select: text; }
  .hint { font-size: 12px; margin: 2px 0 0; line-height: 1.45; }
  .fpfull { font-size: 11px; word-break: break-all; color: var(--muted); user-select: text; -webkit-user-select: text; }

  .toasts { position: fixed; bottom: 16px; right: 16px; display: flex; flex-direction: column; gap: 8px; z-index: 60; max-width: 360px; }
  .toast { padding: 10px 14px; border-radius: 9px; box-shadow: var(--shadow); font-size: 13px; background: var(--panel); border: 1px solid var(--border); border-left: 3px solid var(--accent); animation: up 0.15s ease-out; }
  .toast.error { border-left-color: var(--danger); }
  @keyframes up { from { transform: translateY(6px); opacity: 0; } }

  .fatal { padding: 40px; max-width: 560px; margin: auto; }

  @media (max-width: 800px) {
    .shell { grid-template-columns: 240px 1fr; }
  }
</style>
