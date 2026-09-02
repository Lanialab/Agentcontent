export type Kind = "command" | "async" | "persist" | "external";

export type NodeDef = {
  id: string;
  label: string;
  sub?: string;
  x: number;
  y: number;
  w: number;
  h: number;
  tone?: "ui" | "drive" | "port" | "use" | "domain" | "driven" | "ext";
};

export type EdgeDef = {
  id: string;
  from: string;
  to: string;
  kind: Kind;
  paths: string[];
};

export const VIEW = { w: 1580, h: 1080 };

function n(
  id: string,
  label: string,
  x: number,
  y: number,
  w: number,
  h: number,
  tone: NodeDef["tone"],
  sub?: string,
): NodeDef {
  return { id, label, x, y, w, h, tone, sub };
}

const uiY = 36;
const l1 = 168;
const l2 = 268;
const l3 = 378;
const l4 = 508;
const l5 = 648;
const l6 = 758;
const extX = 1148;

export const nodes: NodeDef[] = [
  n("ui-creators", "Creators & Groups", 48, uiY, 196, 72, "ui", "add · validate · edit · hide · move"),
  n("ui-feed", "Outlier Feed", 258, uiY, 196, 72, "ui", "filters · ranking · five signals"),
  n("ui-ai", "AI Assistant & Ideas", 468, uiY, 210, 72, "ui", "chat · templates · mentions · ideas"),
  n("ui-brand", "Brand Blueprint", 692, uiY, 186, 72, "ui", "Ikigai · positioning · pillars"),
  n("ui-script", "Kịch bản Video", 892, uiY, 196, 72, "ui", "Nhanh · Auto · Sâu · editor"),

  n("http-creator", "Creator HTTP", 48, l1, 150, 56, "drive"),
  n("http-feed", "Feed HTTP", 210, l1, 130, 56, "drive"),
  n("http-ai", "AI HTTP", 352, l1, 120, 56, "drive"),
  n("http-brand", "Brand HTTP", 484, l1, 130, 56, "drive"),
  n("http-script", "Script HTTP", 626, l1, 130, 56, "drive"),
  n("cli-job", "CLI / Job entry", 780, l1, 160, 56, "drive", "every 60 min"),

  n("port-creator", "CreatorCommandPort", 40, l2, 170, 56, "port"),
  n("port-feed", "FeedQueryPort", 222, l2, 150, 56, "port"),
  n("port-ai-in", "AIWorkflowPort", 384, l2, 150, 56, "port"),
  n("port-brand", "BrandPort", 546, l2, 130, 56, "port"),
  n("port-script", "VideoScriptPort", 688, l2, 150, 56, "port"),
  n("port-scan", "ScheduledScanPort", 850, l2, 170, 56, "port"),

  n("uc-creator", "Creator & Scan", 40, l3, 160, 64, "use"),
  n("uc-feed", "Feed Query", 214, l3, 140, 64, "use"),
  n("uc-ai", "AI Assistant & Idea", 368, l3, 170, 64, "use"),
  n("uc-brand", "Brand", 552, l3, 110, 64, "use"),
  n("uc-script", "Video Script", 676, l3, 130, 64, "use"),
  n("uc-scan", "Scheduled Scan", 820, l3, 160, 64, "use", "ASYNC JOB RUNTIME"),

  n("dom-research", "Research Domain", 48, l4, 230, 88, "domain", "groups · channels · scans\nviews/day vs channel avg"),
  n("dom-intel", "Intelligence Domain", 300, l4, 250, 88, "domain", "chat · templates · mentions · ideas"),
  n("dom-brand", "Brand Blueprint", 572, l4, 220, 88, "domain", "Ikigai · voice · pillars · topics"),
  n("dom-script", "Kịch bản Video", 814, l4, 210, 88, "domain", "Nhanh · Auto · Sâu · export"),

  n("port-research-store", "Research store", 48, l5, 150, 52, "port"),
  n("port-video-source", "VideoSourcePort", 210, l5, 150, 52, "port"),
  n("port-ai", "AIProviderPort", 372, l5, 140, 52, "port"),
  n("port-brand-store", "Brand store", 526, l5, 130, 52, "port"),
  n("port-script-store", "Script store", 668, l5, 130, 52, "port"),
  n("port-export", "FileExportPort", 812, l5, 140, 52, "port"),
  n("port-sync", "SyncExportPort", 964, l5, 130, 52, "port"),

  n("adp-sqlite", "SQLite Store", 48, l6, 160, 64, "driven", "sole source of truth"),
  n("adp-youtube", "YouTube adapter", 228, l6, 150, 64, "driven"),
  n("adp-transcript", "Transcript adapter", 396, l6, 160, 64, "driven"),
  n("adp-ai", "AI adapter", 574, l6, 130, 64, "driven"),
  n("adp-sheets", "Sheets adapter", 722, l6, 140, 64, "driven", "NO APPS SCRIPT"),
  n("adp-keychain", "OS Keychain", 880, l6, 120, 64, "driven"),
  n("adp-export", "File export", 1016, l6, 120, 64, "driven", "md · html · pdf"),

  n("ext-youtube", "YouTube Data API", extX, 168, 390, 70, "ext", "channels · uploads · statistics"),
  n("ext-transcript", "Transcript provider", extX, 268, 390, 70, "ext", "Supadata requested — GoClaw current"),
  n("ext-ai", "AI providers", extX, 368, 390, 70, "ext", "OpenAI-compatible · Claude CLI"),
  n("ext-sheets", "Google Sheets API", extX, 468, 390, 70, "ext", "optional sync · NO APPS SCRIPT"),
  n("ext-files", "Local artifact files", extX, 568, 390, 70, "ext", "Markdown · HTML · PDF"),
  n("ext-scheduler", "Native OS scheduler", extX, 758, 390, 70, "ext", "60 min · skip missed · CLI entry"),
];

export const edges: EdgeDef[] = [
  e("e1", "ui-creators", "http-creator", "command", ["creator", "scan"]),
  e("e2", "ui-feed", "http-feed", "command", ["feed"]),
  e("e3", "ui-ai", "http-ai", "command", ["idea"]),
  e("e4", "ui-brand", "http-brand", "command", ["brand"]),
  e("e5", "ui-script", "http-script", "command", ["script"]),

  e("e6", "http-creator", "port-creator", "command", ["creator", "scan"]),
  e("e7", "http-feed", "port-feed", "command", ["feed"]),
  e("e8", "http-ai", "port-ai-in", "command", ["idea"]),
  e("e9", "http-brand", "port-brand", "command", ["brand"]),
  e("e10", "http-script", "port-script", "command", ["script"]),
  e("e11", "cli-job", "port-scan", "async", ["scan"]),

  e("e12", "port-creator", "uc-creator", "command", ["creator", "scan"]),
  e("e13", "port-feed", "uc-feed", "command", ["feed"]),
  e("e14", "port-ai-in", "uc-ai", "command", ["idea"]),
  e("e15", "port-brand", "uc-brand", "command", ["brand"]),
  e("e16", "port-script", "uc-script", "command", ["script"]),
  e("e17", "port-scan", "uc-scan", "async", ["scan"]),

  e("e18", "uc-creator", "dom-research", "command", ["creator", "scan"]),
  e("e19", "uc-feed", "dom-research", "command", ["feed"]),
  e("e20", "uc-ai", "dom-intel", "command", ["idea"]),
  e("e21", "uc-brand", "dom-brand", "command", ["brand"]),
  e("e22", "uc-script", "dom-script", "command", ["script"]),
  e("e23", "uc-scan", "dom-research", "async", ["scan"]),
  e("e24", "uc-scan", "uc-creator", "async", ["scan"]),

  e("e25", "dom-research", "port-research-store", "persist", ["creator", "feed", "scan"]),
  e("e26", "dom-research", "port-video-source", "external", ["scan", "creator"]),
  e("e27", "dom-intel", "port-ai", "external", ["idea"]),
  e("e28", "dom-brand", "port-brand-store", "persist", ["brand"]),
  e("e29", "dom-script", "port-script-store", "persist", ["script"]),
  e("e30", "dom-script", "port-export", "external", ["script"]),

  e("e31", "port-research-store", "adp-sqlite", "persist", ["creator", "feed", "scan", "idea", "brand", "script"]),
  e("e32", "port-video-source", "adp-youtube", "external", ["scan", "creator"]),
  e("e33", "port-ai", "adp-ai", "external", ["idea", "script"]),
  e("e34", "port-brand-store", "adp-sqlite", "persist", ["brand"]),
  e("e35", "port-script-store", "adp-sqlite", "persist", ["script"]),
  e("e36", "port-export", "adp-export", "external", ["script"]),
  e("e37", "port-sync", "adp-sheets", "external", ["scan"]),

  e("e38", "adp-youtube", "ext-youtube", "external", ["scan", "creator"]),
  e("e39", "adp-transcript", "ext-transcript", "external", ["scan"]),
  e("e40", "adp-ai", "ext-ai", "external", ["idea", "script"]),
  e("e41", "adp-sheets", "ext-sheets", "external", ["scan"]),
  e("e42", "adp-export", "ext-files", "external", ["script"]),
  e("e43", "cli-job", "ext-scheduler", "async", ["scan"]),
  e("e44", "adp-keychain", "ext-scheduler", "external", []),
];

function e(id: string, from: string, to: string, kind: Kind, paths: string[]): EdgeDef {
  return { id, from, to, kind, paths };
}

export function nodeMap(): Record<string, NodeDef> {
  const m: Record<string, NodeDef> = {};
  for (const nd of nodes) m[nd.id] = nd;
  return m;
}

export function pathD(a: NodeDef, b: NodeDef): string {
  const x1 = a.x + a.w / 2;
  const y1 = a.y + a.h;
  const x2 = b.x + b.w / 2;
  const y2 = b.y;
  const horizontal = Math.abs(y2 - y1) < 20 || (b.x > a.x + a.w && Math.abs(a.y - b.y) < 40);
  if (b.x > a.x + a.w - 8 && Math.abs(a.y - b.y) < 90) {
    const sx = a.x + a.w;
    const sy = a.y + a.h / 2;
    const tx = b.x;
    const ty = b.y + b.h / 2;
    const mx = (sx + tx) / 2;
    return `M ${sx} ${sy} C ${mx} ${sy}, ${mx} ${ty}, ${tx} ${ty}`;
  }
  if (horizontal && b.x < a.x) {
    const sx = a.x;
    const sy = a.y + a.h / 2;
    const tx = b.x + b.w;
    const ty = b.y + b.h / 2;
    const mx = (sx + tx) / 2;
    return `M ${sx} ${sy} C ${mx} ${sy}, ${mx} ${ty}, ${tx} ${ty}`;
  }
  const my = (y1 + y2) / 2;
  return `M ${x1} ${y1} C ${x1} ${my}, ${x2} ${my}, ${x2} ${y2}`;
}

export const kindColor: Record<Kind, string> = {
  command: "#3ee0ff",
  async: "#ff4fd8",
  persist: "#3dff9a",
  external: "#ffb020",
};
