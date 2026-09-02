export type Page = "arch" | "creators" | "feed" | "ai" | "brand" | "script";

export type Pulse = {
  type: string;
  at: string;
  path: string;
  kind: string;
  label: string;
  nodes: string[];
};

export type Group = { id: string; name: string; sortOrder: number };
export type Channel = {
  id: string;
  groupId: string;
  youtubeId: string;
  handle: string;
  title: string;
  description: string;
  hidden: boolean;
  validated: boolean;
  notes: string;
};
export type Signals = {
  outlierScore: number;
  velocity: number;
  recencyHours: number;
  baselineGap: number;
  hook: number;
};
export type ScoredVideo = {
  id: string;
  channelId: string;
  youtubeId: string;
  title: string;
  publishedAt: string;
  viewCount: number;
  viewsPerDay: number;
  channelTitle: string;
  channelHandle: string;
  groupName: string;
  channelAvgViewsPerDay: number;
  score: number;
  outlier: boolean;
  signals: Signals;
};
export type Idea = {
  id: string;
  title: string;
  prompt: string;
  body: string;
  mentions: string[];
  templateId: string;
  createdAt: string;
};
export type ChatMessage = { id: string; role: string; content: string; createdAt: string };
export type Template = { id: string; name: string; description: string; prompt: string };
export type Blueprint = {
  love: string;
  goodAt: string;
  worldNeeds: string;
  paidFor: string;
  positioning: string;
  voice: string;
  pillars: string[];
  topics: string[];
  updatedAt: string;
};
export type ScriptMode = "nhanh" | "auto" | "sau";
export type Script = {
  id: string;
  mode: ScriptMode;
  title: string;
  body: string;
  status: string;
  updatedAt: string;
};
export type Meta = {
  name: string;
  outlierMultiplier: number;
  youtubeStub: boolean;
  aiStub: boolean;
  sheetsEnabled: boolean;
  outlierNote: string;
};

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
  });
  if (!res.ok) {
    let msg = res.statusText;
    try {
      const j = (await res.json()) as { error?: string };
      if (j.error) msg = j.error;
    } catch {
      /* ignore */
    }
    throw new Error(msg);
  }
  return (await res.json()) as T;
}

export const api = {
  meta: () => req<Meta>("/api/meta"),
  groups: () => req<Group[]>("/api/groups"),
  createGroup: (name: string) => req<Group>("/api/groups", { method: "POST", body: JSON.stringify({ name }) }),
  channels: (hidden = true) => req<Channel[]>(`/api/channels?hidden=${hidden ? "1" : "0"}`),
  addChannel: (body: { groupId: string; handle: string; notes?: string }) =>
    req<Channel>("/api/channels", { method: "POST", body: JSON.stringify(body) }),
  editChannel: (id: string, body: Partial<Channel>) =>
    req<Channel>(`/api/channels/${id}`, { method: "PATCH", body: JSON.stringify(body) }),
  validateChannel: (id: string) => req<Channel>(`/api/channels/${id}/validate`, { method: "POST" }),
  hideChannel: (id: string, hidden: boolean) =>
    req<Channel>(`/api/channels/${id}/hide`, { method: "POST", body: JSON.stringify({ hidden }) }),
  moveChannel: (id: string, groupId: string) =>
    req<Channel>(`/api/channels/${id}/move`, { method: "POST", body: JSON.stringify({ groupId }) }),
  feed: (q: string) => req<ScoredVideo[]>(`/api/feed${q}`),
  scan: () => req<{ channels: number; videos: number; skipped: boolean; stub: boolean }>("/api/scan", { method: "POST" }),
  templates: () => req<Template[]>("/api/templates"),
  chat: () => req<ChatMessage[]>("/api/chat"),
  sendChat: (body: { message: string; templateId?: string; mentions?: string[] }) =>
    req<ChatMessage>("/api/chat", { method: "POST", body: JSON.stringify(body) }),
  ideas: () => req<Idea[]>("/api/ideas"),
  generateIdea: (body: { prompt: string; templateId?: string; mentions?: string[] }) =>
    req<Idea>("/api/ideas", { method: "POST", body: JSON.stringify(body) }),
  brand: () => req<Blueprint>("/api/brand"),
  saveBrand: (bp: Blueprint) => req<Blueprint>("/api/brand", { method: "PUT", body: JSON.stringify(bp) }),
  scripts: () => req<Script[]>("/api/scripts"),
  generateScript: (body: { mode: ScriptMode; topic: string; title?: string }) =>
    req<Script>("/api/scripts", { method: "POST", body: JSON.stringify(body) }),
  saveScript: (s: Script) => req<Script>(`/api/scripts/${s.id}`, { method: "PUT", body: JSON.stringify(s) }),
  exportScript: (id: string, format: string) =>
    req<{ path: string }>(`/api/scripts/${id}/export`, { method: "POST", body: JSON.stringify({ format }) }),
};

export function pageLabel(page: Page): string {
  switch (page) {
    case "arch":
      return "Kiến trúc";
    case "creators":
      return "Creators & Groups";
    case "feed":
      return "Outlier Feed";
    case "ai":
      return "AI & Ideas";
    case "brand":
      return "Brand Blueprint";
    case "script":
      return "Kịch bản Video";
    default: {
      const _never: never = page;
      return _never;
    }
  }
}
