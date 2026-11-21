const en = {
  meta: {
    title: 'PatchDrop: share files over the office cable',
    description:
      'Free, open-source file sharing between Mac and Windows over your wired office network. No internet, no account. Beta.',
  },
  beta: {
    label: 'Beta',
    text: 'Tested in our office on Mac and Windows.',
    ask: 'Something not working?',
    link: 'Tell us on GitHub',
    dismiss: 'Hide this notice',
  },
  nav: {
    how: 'How it works',
    features: 'Features',
    download: 'Download',
    faq: 'FAQ',
    github: 'GitHub',
    language: 'Language',
    home: 'PatchDrop home',
    main: 'Main',
  },
  hero: {
    titleLead: 'Share files over the office cable.',
    titleAccent: 'Just drag and drop.',
    lead: 'The easiest way to move files between Mac and Windows computers on your office network. No internet, no cloud, no account.',
    downloadMac: 'Download for macOS',
    downloadWindows: 'Download for Windows',
    star: 'Star on GitHub',
    starsLabel: 'GitHub stars',
    downloads: 'downloads',
    alsoWindows: 'Also for Windows',
    alsoMac: 'Also for macOS',
    otherNote: 'For macOS and Windows',
    free: 'Free & open source',
    beta: 'beta',
    mobileNote: 'PatchDrop runs on Mac and Windows. Open this page on your computer to download it.',
    sendToComputer: 'Send link to my computer',
    screenshotAlt:
      'The PatchDrop window on a Mac: two other office Macs, Accounts-Mac and Design-Mac, in the device list next to a drop area for files.',
  },
  why: {
    eyebrow: 'Why PatchDrop',
    title: 'Made for offices where the computers are on cable.',
    paragraphs: [
      'In many offices every desk is wired: each computer is plugged into the network with an Ethernet cable. That is great for speed, but file-sharing apps designed around Wi-Fi often can’t find each other there. So files still travel by USB stick or by email to the next desk.',
      'PatchDrop announces itself on every wired network adapter of your computer, so the computers in your office find each other within seconds. If your network blocks that too, add a computer by its IP address or scan the network.',
      'We like LocalSend a lot, and on Wi-Fi it’s a great choice. PatchDrop isn’t trying to replace it. It covers the one case that kept tripping us up: the wired office network.',
    ],
  },
  how: {
    title: 'How it works',
    sub: 'Three steps. No accounts, no setup.',
    steps: [
      { title: 'Install on both computers', text: 'Download PatchDrop on the computer that sends and on the one that receives.' },
      { title: 'Drop your files on the window', text: 'Drag files or whole folders onto PatchDrop, or pick them with a click.' },
      { title: 'Pick the other computer and Send', text: 'It shows up in the list on its own. The other side accepts, and the files arrive.' },
    ],
  },
  features: {
    title: 'Features',
    sub: 'Small app, does one job well.',
    items: [
      { title: 'Full cable speed', text: 'Files go straight across your office network, as fast as the cable allows.' },
      { title: 'Encrypted, verified devices', text: 'Every transfer is encrypted with TLS 1.3. Each computer has its own certificate, and PatchDrop won’t send if it doesn’t match the one it expects.' },
      { title: 'No internet, no cloud', text: 'Nothing leaves your network. It even works when the internet is down.' },
      { title: 'Whole folders', text: 'Send a folder and it arrives with all its subfolders in place.' },
      { title: 'Mac + Windows', text: 'Intel and Apple Silicon Macs, Windows 10 and 11, in any direction.' },
      { title: 'Free, MIT, no ads', text: 'Open source under the MIT license. No ads, no account, no paid tier.' },
    ],
  },
  download: {
    title: 'Download PatchDrop',
    version: 'Version',
    mac: {
      name: 'macOS',
      req: 'macOS 12 or later · Intel and Apple Silicon',
      primary: 'Download .dmg',
      secondary: 'Universal .zip',
    },
    windows: {
      name: 'Windows',
      req: 'Windows 10 and 11 · 64-bit',
      primary: 'Download installer',
      secondary: 'Portable .zip',
    },
    or: 'or',
    macTitle: 'First launch on a Mac',
    macIntro: 'PatchDrop isn’t notarized by Apple yet, so macOS blocks it the first time. To allow it:',
    macSteps: [
      'Open the .dmg and drag PatchDrop into Applications.',
      'Open PatchDrop once. When macOS blocks it, click Done.',
      'Go to System Settings → Privacy & Security.',
      'Scroll down, click Open Anyway next to PatchDrop and confirm. Allow Local Network access when asked.',
    ],
    macFigure: 'Privacy & Security',
    macFigureText: '“PatchDrop” was blocked to protect your Mac.',
    macFigureButton: 'Open Anyway',
    winTitle: 'First launch on Windows',
    winIntro: 'The installer is new and not signed yet, so SmartScreen may warn you:',
    winSteps: [
      'On “Windows protected your PC”, click More info.',
      'Click Run anyway and follow the installer.',
      'If Windows asks about the firewall, allow PatchDrop on private networks.',
    ],
    releaseNotes: 'Release notes',
    checksums: 'SHA-256 checksums',
  },
  share: { title: 'Share' },
  faq: {
    title: 'Questions',
    items: [
      {
        q: 'Is it really free?',
        a: 'Yes. PatchDrop is free and open source under the MIT license. There are no ads, no paid version and no account.',
      },
      {
        q: 'Does it need the internet?',
        a: 'No. The computers talk to each other directly over your local network, so it keeps working when the office internet is down.',
      },
      {
        q: 'Is it safe?',
        a: 'Transfers are encrypted with TLS 1.3, and each computer has its own certificate. Before sending, PatchDrop checks that the other computer’s certificate matches the one it expects, and stops if it doesn’t. Nothing is accepted until the person at the other computer clicks Accept, unless they chose to always accept from you.',
      },
      {
        q: 'The computers don’t see each other. What now?',
        a: 'Check the firewall first and allow PatchDrop on both computers. On Windows, set the Ethernet connection to Private (Settings → Network & internet → Ethernet). Still nothing? Type the other computer’s IP address into Add by IP, or click Scan network. PatchDrop uses UDP port 47820 to find computers and TCP port 47821 for transfers, so both need to be open.',
      },
      {
        q: 'Why not just use LocalSend?',
        a: 'LocalSend is a great app, and on Wi-Fi we’d recommend it too. On wired office networks, though, its device discovery often fails to find the other computers. PatchDrop was built for exactly that case. Use whichever works best on your network.',
      },
      {
        q: 'Where do received files go?',
        a: 'Into a PatchDrop folder inside Downloads. You can pick another folder in the app’s settings. Files with the same name are never overwritten.',
      },
      {
        q: 'What does beta mean here?',
        a: 'A few colleagues have tested PatchDrop in our office on Mac and Windows, but not on every device or system version yet. If something breaks, please open an issue on GitHub. It really helps.',
      },
      {
        q: 'Does this page track me?',
        a: 'Only anonymous visit counts, with GoatCounter. No cookies, no personal data.',
      },
    ],
  },
  madeBy: {
    eyebrow: 'Made by',
    title: 'Built by Selçuk Engin · Engin Sistem',
    bio: 'Selçuk Engin is the founder of Engin Sistem, an Istanbul software company that has been building B2B and custom business systems since 2003. PatchDrop started as a fix for his own office’s wired network. Now it’s free for everyone.',
    photoAlt: 'Selçuk Engin',
    personLinks: 'Follow Selçuk',
    brandLinks: 'Follow Engin Sistem',
    website: 'Visit enginsistem.com',
  },
  footer: {
    copyright: '© 2026 Engin Tasarım Ltd. Şti.',
    github: 'GitHub',
    bug: 'Report a bug',
    releases: 'Releases',
    license: 'MIT License',
    links: 'Project links',
  },
  notFound: {
    title: 'Page not found',
    description: 'That page does not exist, or it has moved. Head back to the PatchDrop home page.',
    back: 'Back to home',
  },
}

export default en
