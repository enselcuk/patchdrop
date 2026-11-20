export interface Peer {
  id: string
  name: string
  os: string
  ip: string
  port: number
  fp: string
  link: string // wired | wifi | ''
  source: string // lan | manual | scan
  online: boolean
  favorite: boolean
  autoAccept: boolean
}

export interface Transfer {
  id: string
  direction: 'send' | 'receive'
  peerName: string
  label: string
  files: number
  total: number
  done: number
  speed: number
  status: 'waiting' | 'active' | 'done' | 'declined' | 'failed' | 'cancelled'
  error?: string
  saved?: string[]
  startedAt: number
}

export interface Self {
  id: string
  name: string
  os: string
  port: number
  fp: string
  ips: string[]
  downloadDir: string
}

export interface FileMeta {
  id: string
  name: string
  size: number
}

export interface Offer {
  id: string
  sender: { id: string; name: string; os: string; fp: string }
  senderIp: string
  files: FileMeta[]
  totalBytes: number
}
