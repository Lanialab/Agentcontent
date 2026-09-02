import { useEffect, useState } from "react";
import { api, type Script, type ScriptMode } from "../api";

export function ScriptsPage({ onPulse }: { onPulse: () => void }) {
  const [items, setItems] = useState<Script[]>([]);
  const [active, setActive] = useState<Script | null>(null);
  const [mode, setMode] = useState<ScriptMode>("nhanh");
  const [topic, setTopic] = useState("Outlier 2.5x tuần này");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  async function load(selectId?: string) {
    const list = await api.scripts();
    setItems(list);
    if (selectId) setActive(list.find((s) => s.id === selectId) ?? list[0] ?? null);
    else if (!active && list[0]) setActive(list[0]);
  }

  useEffect(() => {
    load().catch((e: Error) => setErr(e.message));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  function modeLabel(m: ScriptMode): string {
    switch (m) {
      case "nhanh":
        return "Nhanh";
      case "auto":
        return "Auto";
      case "sau":
        return "Sâu";
      default: {
        const _n: never = m;
        return _n;
      }
    }
  }

  return (
    <div className="page">
      <header className="page-head">
        <div>
          <p className="kicker">Nhanh · Auto · Sâu</p>
          <h1>Kịch bản Video</h1>
          <p className="lede">Draft, edit, export markdown / HTML / PDF stub. Generate pulses the script + AI wires.</p>
        </div>
      </header>
      {err && <p className="err">{err}</p>}
      <div className="row">
        {(["nhanh", "auto", "sau"] as ScriptMode[]).map((m) => (
          <button key={m} className={mode === m ? "on" : ""} onClick={() => setMode(m)}>
            {modeLabel(m)}
          </button>
        ))}
        <input className="grow" value={topic} onChange={(e) => setTopic(e.target.value)} />
        <button
          className="primary"
          disabled={busy}
          onClick={() => {
            setBusy(true);
            api
              .generateScript({ mode, topic })
              .then((s) => {
                onPulse();
                return load(s.id);
              })
              .catch((e: Error) => setErr(e.message))
              .finally(() => setBusy(false));
          }}
        >
          Generate
        </button>
      </div>
      <div className="split">
        <ul className="script-nav">
          {items.map((s) => (
            <li key={s.id}>
              <button className={active?.id === s.id ? "on" : ""} onClick={() => setActive(s)}>
                <em>{modeLabel(s.mode)}</em>
                {s.title}
              </button>
            </li>
          ))}
        </ul>
        {active && (
          <section className="panel grow editor">
            <input value={active.title} onChange={(e) => setActive({ ...active, title: e.target.value })} />
            <textarea value={active.body} onChange={(e) => setActive({ ...active, body: e.target.value })} />
            <div className="actions">
              <button
                className="primary"
                onClick={() => {
                  api
                    .saveScript(active)
                    .then((s) => {
                      setActive(s);
                      onPulse();
                      return load(s.id);
                    })
                    .catch((e: Error) => setErr(e.message));
                }}
              >
                Save
              </button>
              {(["md", "html", "pdf"] as const).map((fmt) => (
                <button
                  key={fmt}
                  onClick={() => {
                    api
                      .exportScript(active.id, fmt)
                      .then((r) => alert(`Exported ${r.path}`))
                      .catch((e: Error) => setErr(e.message));
                  }}
                >
                  Export {fmt}
                </button>
              ))}
            </div>
          </section>
        )}
      </div>
    </div>
  );
}
