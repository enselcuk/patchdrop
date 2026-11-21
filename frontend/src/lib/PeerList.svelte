<script lang="ts">
  import Icon from './Icon.svelte'
  import type { Peer } from '../types'
  import { osLabel } from '../format'

  let {
    peers,
    selected = $bindable(''),
    onfavorite,
  }: { peers: Peer[]; selected: string; onfavorite: (p: Peer) => void } = $props()

  function osIcon(os: string) {
    return os === 'windows' ? 'windows' : os === 'darwin' ? 'laptop' : 'monitor'
  }
</script>

{#if peers.length === 0}
  <div class="empty">
    <div class="pulse"></div>
    <p><strong>Looking for devices…</strong></p>
    <p class="muted">Open PatchDrop on another computer on this network. If nothing shows up, add it by IP below.</p>
  </div>
{:else}
  <ul>
    {#each peers as p (p.id)}
      <li>
        <button
          class="peer"
          class:sel={selected === p.id}
          class:off={!p.online}
          onclick={() => (selected = p.id)}
          title={`${p.ip}:${p.port}\nFingerprint ${p.fp.slice(0, 16)}…`}
        >
          <span class="os"><Icon name={osIcon(p.os)} size={20} /></span>
          <span class="info">
            <span class="name">{p.name}</span>
            <span class="meta">
              {#if !p.online}
                offline
              {:else}
                {osLabel(p.os)} · {p.ip}
              {/if}
            </span>
          </span>
          {#if p.online && p.link === 'wired'}
            <span class="badge" title="Reached over a wired connection"><Icon name="cable" size={13} /></span>
          {:else if p.online && p.link === 'wifi'}
            <span class="badge" title="Reached over Wi-Fi"><Icon name="wifi" size={13} /></span>
          {/if}
        </button>
        <button
          class="ghost fav"
          class:on={p.favorite}
          title={p.favorite ? 'Remove from favorites' : 'Add to favorites (remembered even when discovery fails)'}
          aria-label="Toggle favorite"
          onclick={() => onfavorite(p)}><Icon name="star" size={15} /></button
        >
      </li>
    {/each}
  </ul>
{/if}

<style>
  ul { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 4px; }
  li { position: relative; }
  .peer {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 10px;
    text-align: left;
    border: 1px solid transparent;
    background: transparent;
    padding: 9px 36px 9px 10px;
    border-radius: 9px;
  }
  .peer:hover { background: var(--panel-2); }
  .peer.sel { background: var(--accent-soft); border-color: var(--accent); }
  .peer.off { opacity: 0.55; }
  .os {
    width: 34px; height: 34px; border-radius: 8px; display: grid; place-items: center;
    background: var(--panel-2); border: 1px solid var(--border); color: var(--muted); flex-shrink: 0;
  }
  .sel .os { color: var(--accent); border-color: var(--accent); }
  .info { display: flex; flex-direction: column; min-width: 0; flex: 1; }
  .name { font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .meta { color: var(--muted); font-size: 12px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .badge { color: var(--muted); display: flex; }
  .fav { position: absolute; right: 4px; top: 50%; transform: translateY(-50%); color: var(--muted); opacity: 0.5; }
  .fav.on { color: #e3a008; opacity: 1; }
  .fav.on :global(path) { fill: currentColor; }
  li:hover .fav { opacity: 1; }
  .empty { text-align: center; padding: 28px 10px; }
  .empty p { margin: 6px 0; font-size: 13px; }
  .pulse {
    width: 14px; height: 14px; border-radius: 50%; background: var(--accent); margin: 0 auto 14px;
    animation: pulse 1.6s ease-out infinite;
  }
  @keyframes pulse {
    0% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--accent) 60%, transparent); }
    100% { box-shadow: 0 0 0 18px transparent; }
  }
</style>
