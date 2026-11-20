<script lang="ts">
  let { name, size = 16 }: { name: string; size?: number } = $props()

  const paths: Record<string, string> = {
    laptop: 'M4 5h16v10H4zM2 19h20',
    windows: 'M3 5.5 10 4.5v7H3zM12 4.2 21 3v8.5h-9zM3 13h7v7l-7-1zM12 13h9v8l-9-1.2z',
    monitor: 'M3 4h18v12H3zM8 20h8M12 16v4',
    refresh: 'M20 11a8 8 0 1 0-2.3 5.7M20 4v7h-7',
    plus: 'M12 5v14M5 12h14',
    radar: 'M12 12m-9 0a9 9 0 1 0 18 0a9 9 0 1 0-18 0M12 12m-5 0a5 5 0 1 0 10 0a5 5 0 1 0-10 0M12 12l6-6',
    star: 'm12 3 2.7 5.6 6.1.9-4.4 4.3 1 6.1L12 17l-5.4 2.9 1-6.1-4.4-4.3 6.1-.9z',
    upload: 'M12 16V4M7 9l5-5 5 5M4 16v4h16v-4',
    down: 'M12 4v12M7 11l5 5 5-5M4 20h16',
    up: 'M12 20V8M7 13l5-5 5 5M4 4h16',
    file: 'M14 3H6v18h12V7zM14 3v4h4',
    folder: 'M3 6h6l2 2h10v11H3z',
    x: 'M6 6l12 12M18 6 6 18',
    gear: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM19.4 13a7.6 7.6 0 0 0 0-2l2-1.6-2-3.4-2.4.9a7.4 7.4 0 0 0-1.7-1L15 3h-4l-.4 2.6a7.4 7.4 0 0 0-1.7 1l-2.4-.9-2 3.4L6.6 11a7.6 7.6 0 0 0 0 2l-2 1.6 2 3.4 2.4-.9a7.4 7.4 0 0 0 1.7 1L11 21h4l.4-2.6a7.4 7.4 0 0 0 1.7-1l2.4.9 2-3.4z',
    cable: 'M7 3v5M17 3v5M5 8h14v4a7 7 0 0 1-14 0zM12 19v3',
    wifi: 'M2 9a15 15 0 0 1 20 0M5 12.5a10 10 0 0 1 14 0M8.5 16a5 5 0 0 1 7 0M12 19.5h.01',
    open: 'M14 4h6v6M20 4l-9 9M18 14v6H4V6h6',
    check: 'm5 12 5 5 9-10',
  }
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="1.8"
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"
>
  <path d={paths[name] ?? ''} />
</svg>
