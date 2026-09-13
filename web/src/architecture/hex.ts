export type Kind = "command" | "async" | "persist" | "external";
export type Tone = "ui" | "drive" | "port" | "use" | "domain" | "driven" | "ext";

export type Chip = {
  id: string;
  label: string;
  sub?: string;
  tone: Tone;
  paths: string[];
};

/** Flat-top hex: width = 2R, height = √3 R. Extra viewBox pad keeps layer 6 on-screen. */
export const HEX = {
  w: 760,
  h: 740,
  cx: 380,
  cy: 372,
  r: 328,
};

export const SQRT3 = Math.sqrt(3);

export function hexVertices(cx: number, cy: number, r: number): [number, number][] {
  const pts: [number, number][] = [];
  for (let i = 0; i < 6; i++) {
    const ang = (Math.PI / 180) * (60 * i);
    pts.push([cx + r * Math.cos(ang), cy + r * Math.sin(ang)]);
  }
  return pts;
}

export function hexPointsAttr(cx: number, cy: number, r: number): string {
  return hexVertices(cx, cy, r)
    .map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`)
    .join(" ");
}

/** Horizontal span of a flat-top hex at a given y. */
export function hexXRange(cx: number, cy: number, r: number, y: number): [number, number] | null {
  const pts = hexVertices(cx, cy, r);
  const xs: number[] = [];
  for (let i = 0; i < 6; i++) {
    const [x1, y1] = pts[i];
    const [x2, y2] = pts[(i + 1) % 6];
    const minY = Math.min(y1, y2);
    const maxY = Math.max(y1, y2);
    if (y < minY - 0.01 || y > maxY + 0.01) continue;
    if (Math.abs(y2 - y1) < 0.01) {
      xs.push(x1, x2);
      continue;
    }
    const t = (y - y1) / (y2 - y1);
    if (t >= -0.01 && t <= 1.01) xs.push(x1 + t * (x2 - x1));
  }
  if (xs.length < 2) return null;
  return [Math.min(...xs), Math.max(...xs)];
}

/** Polygon for a horizontal slice of the hex between y0 and y1. */
export function hexBandPolygon(cx: number, cy: number, r: number, y0: number, y1: number): string {
  const top = hexXRange(cx, cy, r, y0);
  const bot = hexXRange(cx, cy, r, y1);
  if (!top || !bot) return "";
  const pts = hexVertices(cx, cy, r);
  const mid: [number, number][] = [];
  for (const [x, y] of pts) {
    if (y > y0 + 0.5 && y < y1 - 0.5) mid.push([x, y]);
  }
  mid.sort((a, b) => a[0] - b[0]);
  const leftMid = mid.filter((p) => p[0] < cx);
  const rightMid = mid.filter((p) => p[0] >= cx);
  const ring: [number, number][] = [[top[0], y0], [top[1], y0], ...rightMid, [bot[1], y1], [bot[0], y1], ...leftMid.reverse()];
  return ring.map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`).join(" ");
}

export type LayerId = "drive" | "inbound" | "use" | "domain" | "outbound" | "driven";

export type LayerDef = {
  id: LayerId;
  n: number;
  title: string;
  hint: string;
  chips: Chip[];
};

export const layers: LayerDef[] = [
  {
    id: "drive",
    n: 1,
    title: "Driving adapters",
    hint: "HTTP handlers + CLI / job entry",
    chips: [
      { id: "http-creator", label: "Creator HTTP", tone: "drive", paths: ["creator", "scan"] },
      { id: "http-feed", label: "Feed HTTP", tone: "drive", paths: ["feed"] },
      { id: "http-ai", label: "AI HTTP", tone: "drive", paths: ["idea"] },
      { id: "http-brand", label: "Brand HTTP", tone: "drive", paths: ["brand"] },
      { id: "http-script", label: "Script HTTP", tone: "drive", paths: ["script"] },
      { id: "cli-job", label: "CLI / Job", sub: "60 min", tone: "drive", paths: ["scan"] },
    ],
  },
  {
    id: "inbound",
    n: 2,
    title: "Inbound ports",
    hint: "interfaces owned by the core",
    chips: [
      { id: "port-creator", label: "CreatorCommandPort", tone: "port", paths: ["creator", "scan"] },
      { id: "port-feed", label: "FeedQueryPort", tone: "port", paths: ["feed"] },
      { id: "port-ai-in", label: "AIWorkflowPort", tone: "port", paths: ["idea"] },
      { id: "port-brand", label: "BrandPort", tone: "port", paths: ["brand"] },
      { id: "port-script", label: "VideoScriptPort", tone: "port", paths: ["script"] },
      { id: "port-scan", label: "ScheduledScanPort", tone: "port", paths: ["scan"] },
    ],
  },
  {
    id: "use",
    n: 3,
    title: "Application use cases",
    hint: "ASYNC JOB RUNTIME · lease · progress · cancel",
    chips: [
      { id: "uc-creator", label: "Creator & Scan", tone: "use", paths: ["creator", "scan"] },
      { id: "uc-feed", label: "Feed Query", tone: "use", paths: ["feed"] },
      { id: "uc-ai", label: "AI Assistant & Idea", tone: "use", paths: ["idea"] },
      { id: "uc-brand", label: "Brand", tone: "use", paths: ["brand"] },
      { id: "uc-script", label: "Video Script", tone: "use", paths: ["script"] },
      { id: "uc-scan", label: "Scheduled Scan", tone: "use", paths: ["scan"] },
    ],
  },
  {
    id: "domain",
    n: 4,
    title: "Domain models",
    hint: "pure rules · views/day vs channel avg · *100 is diagram seed",
    chips: [
      { id: "dom-research", label: "Research", sub: "groups · channels · scans", tone: "domain", paths: ["creator", "feed", "scan"] },
      { id: "dom-intel", label: "Intelligence", sub: "chat · templates · ideas", tone: "domain", paths: ["idea"] },
      { id: "dom-brand", label: "Brand Blueprint", sub: "Ikigai · pillars · voice", tone: "domain", paths: ["brand"] },
      { id: "dom-script", label: "Kịch bản Video", sub: "Nhanh · Auto · Sâu", tone: "domain", paths: ["script"] },
    ],
  },
  {
    id: "outbound",
    n: 5,
    title: "Outbound ports",
    hint: "interfaces owned by the core",
    chips: [
      { id: "port-research-store", label: "Research", tone: "port", paths: ["creator", "feed", "scan"] },
      { id: "port-video-source", label: "VideoSourcePort", tone: "port", paths: ["scan", "creator"] },
      { id: "port-ai", label: "AIProviderPort", tone: "port", paths: ["idea", "script"] },
      { id: "port-brand-store", label: "Brand", tone: "port", paths: ["brand"] },
      { id: "port-script-store", label: "Video Script", tone: "port", paths: ["script"] },
      { id: "port-export", label: "FileExportPort", tone: "port", paths: ["script"] },
      { id: "port-sync", label: "SyncExportPort", tone: "port", paths: ["scan"] },
    ],
  },
  {
    id: "driven",
    n: 6,
    title: "Driven adapters",
    hint: "inside Go process",
    chips: [
      { id: "adp-sqlite", label: "SQLite", sub: "sole SoT", tone: "driven", paths: ["creator", "feed", "scan", "idea", "brand", "script"] },
      { id: "adp-youtube", label: "YouTube", tone: "driven", paths: ["scan", "creator"] },
      { id: "adp-transcript", label: "Transcript", tone: "driven", paths: ["scan"] },
      { id: "adp-ai", label: "AI adapter", tone: "driven", paths: ["idea", "script"] },
      { id: "adp-sheets", label: "Sheets", sub: "NO APPS SCRIPT", tone: "driven", paths: ["scan"] },
      { id: "adp-keychain", label: "OS Keychain", tone: "driven", paths: [] },
      { id: "adp-export", label: "File export", sub: "md · html · pdf", tone: "driven", paths: ["script"] },
    ],
  },
];

export const uiModules: Chip[] = [
  { id: "ui-creators", label: "Creators & Groups", sub: "add · validate · edit · hide · move", tone: "ui", paths: ["creator", "scan"] },
  { id: "ui-feed", label: "Outlier Feed", sub: "filters · ranking · five signals", tone: "ui", paths: ["feed"] },
  { id: "ui-ai", label: "AI Assistant & Ideas", sub: "chat · templates · mentions · ideas", tone: "ui", paths: ["idea"] },
  { id: "ui-brand", label: "Brand Blueprint", sub: "Ikigai · positioning · pillars", tone: "ui", paths: ["brand"] },
  { id: "ui-script", label: "Kịch bản Video", sub: "Nhanh · Auto · Sâu · editor", tone: "ui", paths: ["script"] },
];

export const extSystems: Chip[] = [
  { id: "ext-youtube", label: "YouTube Data API", sub: "channels · uploads · statistics", tone: "ext", paths: ["scan", "creator"] },
  { id: "ext-transcript", label: "Transcript provider", sub: "Supadata requested — GoClaw current", tone: "ext", paths: ["scan"] },
  { id: "ext-ai", label: "AI providers", sub: "OpenAI-compatible · Claude CLI", tone: "ext", paths: ["idea", "script"] },
  { id: "ext-sheets", label: "Google Sheets API", sub: "optional sync · NO APPS SCRIPT", tone: "ext", paths: ["scan"] },
  { id: "ext-files", label: "Local artifact files", sub: "Markdown · HTML · PDF", tone: "ext", paths: ["script"] },
  { id: "ext-scheduler", label: "Native OS scheduler", sub: "60 min · skip missed · CLI entry", tone: "ext", paths: ["scan"] },
];

export type LaidChip = Chip & { x: number; y: number; w: number; h: number };

export type LaidLayer = Omit<LayerDef, "chips"> & {
  y0: number;
  y1: number;
  chips: LaidChip[];
};

function placeRow(chips: Chip[], y: number, ch: number, innerL: number, avail: number): LaidChip[] {
  const gap = 6;
  const count = chips.length;
  const cw = Math.min(124, Math.max(70, (avail - gap * (count - 1)) / count));
  const total = count * cw + gap * (count - 1);
  let x = innerL + Math.max(0, (avail - total) / 2);
  return chips.map((c) => {
    const item: LaidChip = { ...c, x, y, w: cw, h: ch };
    x += cw + gap;
    return item;
  });
}

export function layoutHex(): LaidLayer[] {
  const { cx, cy, r } = HEX;
  const top = cy - (SQRT3 / 2) * r;
  const bot = cy + (SQRT3 / 2) * r;
  const h = bot - top;
  const weights = [0.155, 0.135, 0.15, 0.2, 0.145, 0.215];
  let y = top + 16;
  const out: LaidLayer[] = [];
  for (let i = 0; i < layers.length; i++) {
    const layer = layers[i];
    const y0 = y;
    const y1 = i === layers.length - 1 ? bot - 14 : y + h * weights[i];
    const midY = (y0 + y1) / 2;
    const span = hexXRange(cx, cy, r - 8, midY) ?? [cx - 200, cx + 200];
    const pad = 18;
    const innerL = span[0] + pad;
    const innerR = span[1] - pad;
    const avail = Math.max(140, innerR - innerL);
    const twoRow = layer.id === "drive" || layer.id === "driven";
    const chips = layer.chips;
    const titleH = 14;
    let laid: LaidChip[];
    if (twoRow && chips.length > 4) {
      const split = Math.ceil(chips.length / 2);
      const ch = Math.min(34, (y1 - y0 - titleH - 10) / 2);
      const yA = y0 + titleH;
      const yB = yA + ch + 4;
      laid = [...placeRow(chips.slice(0, split), yA, ch, innerL, avail), ...placeRow(chips.slice(split), yB, ch, innerL, avail)];
    } else {
      const ch = Math.min(44, y1 - y0 - titleH - 10);
      laid = placeRow(chips, y0 + titleH + (y1 - y0 - titleH - ch) / 2, ch, innerL, avail);
    }
    out.push({ ...layer, y0, y1, chips: laid });
    y = y1;
  }
  return out;
}

export type Wire = {
  id: string;
  d: string;
  kind: Kind;
  paths: string[];
  from: string;
  to: string;
};

function bottom(c: LaidChip): [number, number] {
  return [c.x + c.w / 2, c.y + c.h];
}
function topPt(c: LaidChip): [number, number] {
  return [c.x + c.w / 2, c.y];
}
function curve(a: [number, number], b: [number, number]): string {
  const my = (a[1] + b[1]) / 2;
  return `M ${a[0].toFixed(1)} ${a[1].toFixed(1)} C ${a[0].toFixed(1)} ${my.toFixed(1)}, ${b[0].toFixed(1)} ${my.toFixed(1)}, ${b[0].toFixed(1)} ${b[1].toFixed(1)}`;
}

function byId(laid: LaidLayer[]): Record<string, LaidChip> {
  const m: Record<string, LaidChip> = {};
  for (const layer of laid) {
    for (const c of layer.chips) m[c.id] = c;
  }
  return m;
}

export function hexWires(laid: LaidLayer[]): Wire[] {
  const m = byId(laid);
  const pairs: [string, string, Kind, string[]][] = [
    ["http-creator", "port-creator", "command", ["creator", "scan"]],
    ["http-feed", "port-feed", "command", ["feed"]],
    ["http-ai", "port-ai-in", "command", ["idea"]],
    ["http-brand", "port-brand", "command", ["brand"]],
    ["http-script", "port-script", "command", ["script"]],
    ["cli-job", "port-scan", "async", ["scan"]],
    ["port-creator", "uc-creator", "command", ["creator", "scan"]],
    ["port-feed", "uc-feed", "command", ["feed"]],
    ["port-ai-in", "uc-ai", "command", ["idea"]],
    ["port-brand", "uc-brand", "command", ["brand"]],
    ["port-script", "uc-script", "command", ["script"]],
    ["port-scan", "uc-scan", "async", ["scan"]],
    ["uc-creator", "dom-research", "command", ["creator", "scan"]],
    ["uc-feed", "dom-research", "command", ["feed"]],
    ["uc-ai", "dom-intel", "command", ["idea"]],
    ["uc-brand", "dom-brand", "command", ["brand"]],
    ["uc-script", "dom-script", "command", ["script"]],
    ["uc-scan", "dom-research", "async", ["scan"]],
    ["dom-research", "port-research-store", "persist", ["creator", "feed", "scan"]],
    ["dom-research", "port-video-source", "external", ["scan", "creator"]],
    ["dom-intel", "port-ai", "external", ["idea"]],
    ["dom-brand", "port-brand-store", "persist", ["brand"]],
    ["dom-script", "port-script-store", "persist", ["script"]],
    ["dom-script", "port-export", "external", ["script"]],
    ["port-research-store", "adp-sqlite", "persist", ["creator", "feed", "scan", "idea", "brand", "script"]],
    ["port-video-source", "adp-youtube", "external", ["scan", "creator"]],
    ["port-ai", "adp-ai", "external", ["idea", "script"]],
    ["port-brand-store", "adp-sqlite", "persist", ["brand"]],
    ["port-script-store", "adp-sqlite", "persist", ["script"]],
    ["port-export", "adp-export", "external", ["script"]],
    ["port-sync", "adp-sheets", "external", ["scan"]],
  ];
  const wires: Wire[] = [];
  pairs.forEach(([from, to, kind, paths], i) => {
    const a = m[from];
    const b = m[to];
    if (!a || !b) return;
    wires.push({ id: `w${i}`, from, to, kind, paths, d: curve(bottom(a), topPt(b)) });
  });
  return wires;
}

export const kindColor: Record<Kind, string> = {
  command: "#3ee0ff",
  async: "#ff4fd8",
  persist: "#3dff9a",
  external: "#ffb020",
};

export const uiToDrive: { ui: string; drive: string; paths: string[] }[] = [
  { ui: "ui-creators", drive: "http-creator", paths: ["creator", "scan"] },
  { ui: "ui-feed", drive: "http-feed", paths: ["feed"] },
  { ui: "ui-ai", drive: "http-ai", paths: ["idea"] },
  { ui: "ui-brand", drive: "http-brand", paths: ["brand"] },
  { ui: "ui-script", drive: "http-script", paths: ["script"] },
];

export const driveToExt: { adp: string; ext: string; kind: Kind; paths: string[] }[] = [
  { adp: "adp-youtube", ext: "ext-youtube", kind: "external", paths: ["scan", "creator"] },
  { adp: "adp-transcript", ext: "ext-transcript", kind: "external", paths: ["scan"] },
  { adp: "adp-ai", ext: "ext-ai", kind: "external", paths: ["idea", "script"] },
  { adp: "adp-sheets", ext: "ext-sheets", kind: "external", paths: ["scan"] },
  { adp: "adp-export", ext: "ext-files", kind: "external", paths: ["script"] },
  { adp: "cli-job", ext: "ext-scheduler", kind: "async", paths: ["scan"] },
];
