<script lang="ts">
  import Icon from './Icon.svelte'
  import type { Transfer } from '../types'
  import { bytes, eta } from '../format'

  let {
    transfers,
    oncancel,
    onreveal,
    onclear,
  }: {
    transfers: Transfer[]
    oncancel: (id: string) => void
    onreveal: (path: string) => void
    onclear: () => void
  } = $props()

  const finished = (t: Transfer) => ['done', 'declined', 'failed', 'cancelled'].includes(t.status)

  function statusText(t: Transfer) {
    switch (t.status) {
      case 'waiting': return `Waiting for ${t.peerName} to accept…`
      case 'active': {
        const left = eta(t.total - t.done, t.speed)
        return `${bytes(t.done)} of ${bytes(t.total)}${t.speed ? ` · ${bytes(t.speed)}/s` : ''}${left ? ` · ${left}` : ''}`
      }
      case 'done': return `${bytes(t.total)} · ${t.direction === 'send' ? 'sent' : 'received'}`
      case 'declined': return `${t.peerName} declined`
      case 'cancelled': return 'Cancelled'
      case 'failed': return t.error || 'Failed'
    }
  }
  const pct = (t: Transfer) => (t.total > 0 ? (t.done / t.total) * 100 : t.status === 'done' ? 100 : 0)
</script>

<div class="head">
  <h2>Transfers</h2>
  {#if transfers.some(finished)}
    <button class="ghost small" onclick={onclear}>Clear finished</button>
  {/if}
</div>

{#if transfers.length === 0}
  <p class="muted none">Nothing sent or received yet.</p>
{:else}
  <ul>
    {#each transfers as t (t.id)}
      <li class={t.status}>
        <span class="dir" title={t.direction === 'send' ? 'Outgoing' : 'Incoming'}>
          <Icon name={t.direction === 'send' ? 'up' : 'down'} size={16} />
        </span>
        <div class="body">
          <div class="row">
            <span class="label" title={t.label}>{t.label}</span>
            <span class="peer muted">{t.direction === 'send' ? 'to' : 'from'} {t.peerName}{t.files > 1 ? ` · ${t.files} files` : ''}</span>
          </div>
          {#if !finished(t)}
            <div class="bar"><div class="fill" class:indet={t.status === 'waiting'} style:width={`${pct(t)}%`}></div></div>
          {/if}
          <div class="status">{statusText(t)}</div>
        </div>
        <div class="actions">
          {#if !finished(t)}
            <button class="ghost" title="Cancel" aria-label="Cancel" onclick={() => oncancel(t.id)}><Icon name="x" /></button>
          {:else if t.direction === 'receive' && t.status === 'done' && t.saved?.length}
            <button class="ghost" title="Show in folder" aria-label="Show in folder" onclick={() => onreveal(t.saved![0])}><Icon name="open" /></button>
          {/if}
        </div>
      </li>
    {/each}
  </ul>
{/if}

<style>
  .head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
  h2 { font-size: 13px; text-transform: uppercase; letter-spacing: 0.05em; color: var(--muted); margin: 0; }
  .small { font-size: 12px; }
  .none { font-size: 13px; margin: 6px 0; }
  ul { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 6px; }
  li {
    display: flex; gap: 10px; align-items: center;
    background: var(--panel); border: 1px solid var(--border); border-radius: var(--radius);
    padding: 10px 10px 10px 12px;
  }
  .dir { color: var(--accent); display: flex; }
  .failed .dir, .declined .dir { color: var(--danger); }
  .cancelled .dir { color: var(--muted); }
  .done .dir { color: var(--ok); }
  .body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 5px; }
  .row { display: flex; gap: 8px; align-items: baseline; min-width: 0; }
  .label { font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .peer { font-size: 12px; white-space: nowrap; }
  .status { font-size: 12px; color: var(--muted); font-variant-numeric: tabular-nums; }
  .failed .status { color: var(--danger); }
  .bar { height: 5px; background: var(--panel-2); border: 1px solid var(--border); border-radius: 4px; overflow: hidden; }
  .fill { height: 100%; background: var(--accent); transition: width 0.2s linear; }
  .fill.indet { width: 30% !important; animation: slide 1.2s ease-in-out infinite; opacity: 0.6; }
  @keyframes slide { 0% { margin-left: -30%; } 100% { margin-left: 100%; } }
  .actions { display: flex; }
</style>
