import { useEffect, useState } from "react";
import { api, type Group, type ScoredVideo } from "../api";
import type { DemoHandoff } from "../App";

type Props = {
  onRemixIdea: (handoff: DemoHandoff) => void;
};

export function FeedPage({ onRemixIdea }: Props) {
  const [items, setItems] = useState<ScoredVideo[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [q, setQ] = useState("");
  const [groupId, setGroupId] = useState("");
  const [outliersOnly, setOutliersOnly] = useState(true);
  const [err, setErr] = useState("");
  const [scanning, setScanning] = useState(false);

  async function load() {
    const params = new URLSearchParams();
    if (q) params.set("q", q);
    if (groupId) params.set("groupId", groupId);
    if (outliersOnly) params.set("outliers", "1");
    const qs = params.toString() ? `?${params}` : "";
    const [feed, g] = await Promise.all([api.feed(qs), api.groups()]);
    setItems(feed);
    setGroups(g);
  }

  useEffect(() => {
    load().catch((e: Error) => setErr(e.message));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [groupId, outliersOnly]);

  function remixIdea(v: ScoredVideo) {
    onRemixIdea({
      prompt: `Outlier ${v.score.toFixed(2)}x: "${v.title}" của @${v.channelHandle}. Viết 1 idea remix khớp Brand pillars Research & outliers; giữ hook mạnh, góc nhìn riêng.`,
      mentions: `@${v.channelHandle}`,
      templateId: "outlier-remix",
      sourceVideoTitle: v.title,
      sourceChannel: v.channelHandle,
    });
  }

  return (
    <div className="page">
      <header className="page-head">
        <div>
          <p className="kicker">Five signals</p>
          <h1>Outlier Feed</h1>
          <p className="lede">
            Ranking = viewsPerDay ÷ average viewsPerDay of recent same-channel videos. Default 2.5x. The architecture diagram’s *100 is a visual seed note, not the threshold.
          </p>
        </div>
      </header>
      {err && <p className="err">{err}</p>}
      <form
        className="filters"
        onSubmit={(e) => {
          e.preventDefault();
          load().catch((er: Error) => setErr(er.message));
        }}
      >
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Search title / channel" />
        <select value={groupId} onChange={(e) => setGroupId(e.target.value)}>
          <option value="">All groups</option>
          {groups.map((g) => (
            <option key={g.id} value={g.id}>
              {g.name}
            </option>
          ))}
        </select>
        <label className="chk">
          <input type="checkbox" checked={outliersOnly} onChange={(e) => setOutliersOnly(e.target.checked)} />
          Outliers only
        </label>
        <button type="submit">Filter</button>
        <button
          type="button"
          disabled={scanning}
          onClick={() => {
            setScanning(true);
            setErr("");
            api
              .scan()
              .then(() => load())
              .catch((er: Error) => setErr(er.message))
              .finally(() => setScanning(false));
          }}
        >
          {scanning ? "Đang quét…" : "Làm mới feed"}
        </button>
      </form>
      <div className="feed-list">
        {items.map((v) => (
          <article key={v.id} className={`feed-card ${v.outlier ? "outlier" : ""}`}>
            <div className="score">{v.score.toFixed(2)}x</div>
            <div>
              <h3>{v.title}</h3>
              <p className="muted">
                @{v.channelHandle} · {v.groupName} · {Math.round(v.viewCount).toLocaleString()} views
              </p>
              <div className="signals">
                <Signal n="outlier" v={v.signals.outlierScore.toFixed(2)} />
                <Signal n="velocity" v={Math.round(v.signals.velocity).toLocaleString()} />
                <Signal n="recency h" v={v.signals.recencyHours.toFixed(0)} />
                <Signal n="baseline Δ" v={Math.round(v.signals.baselineGap).toLocaleString()} />
                <Signal n="hook" v={v.signals.hook.toFixed(2)} />
              </div>
              <div className="cta">
                <button type="button" className="primary" onClick={() => remixIdea(v)}>
                  Tạo idea
                </button>
              </div>
            </div>
          </article>
        ))}
        {items.length === 0 && <p className="muted">No videos match. Seed data should show 4 outliers at 2.5x.</p>}
      </div>
    </div>
  );
}

function Signal({ n, v }: { n: string; v: string }) {
  return (
    <span className="sig">
      <em>{n}</em>
      {v}
    </span>
  );
}
