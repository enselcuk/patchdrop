<script lang="ts">
  import type { Snippet } from 'svelte'
  let { title, onclose, children }: { title: string; onclose?: () => void; children: Snippet } = $props()

  function key(e: KeyboardEvent) {
    if (e.key === 'Escape' && onclose) onclose()
  }
</script>

<svelte:window onkeydown={key} />

<div class="backdrop">
  <div class="modal" role="dialog" aria-modal="true" aria-label={title}>
    <h3>{title}</h3>
    {@render children()}
  </div>
</div>

<style>
  .backdrop {
    position: fixed; inset: 0; background: rgba(10, 12, 16, 0.45);
    display: grid; place-items: center; z-index: 50; padding: 16px;
    animation: fade 0.12s ease-out;
  }
  .modal {
    background: var(--panel); border: 1px solid var(--border); border-radius: 14px;
    width: min(460px, 100%); max-height: calc(100vh - 32px); overflow: auto;
    padding: 20px; box-shadow: 0 20px 50px rgba(0, 0, 0, 0.25);
  }
  h3 { margin: 0 0 12px; font-size: 16px; }
  @keyframes fade { from { opacity: 0; } }
</style>
