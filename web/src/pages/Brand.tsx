import { useEffect, useState } from "react";
import { api, type Blueprint } from "../api";

const empty: Blueprint = {
  love: "",
  goodAt: "",
  worldNeeds: "",
  paidFor: "",
  positioning: "",
  voice: "",
  pillars: [],
  topics: [],
  updatedAt: "",
};

export function BrandPage({ onPulse }: { onPulse: () => void }) {
  const [bp, setBp] = useState<Blueprint>(empty);
  const [pillars, setPillars] = useState("");
  const [topics, setTopics] = useState("");
  const [err, setErr] = useState("");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    api
      .brand()
      .then((b) => {
        setBp(b);
        setPillars((b.pillars ?? []).join("\n"));
        setTopics((b.topics ?? []).join("\n"));
      })
      .catch((e: Error) => setErr(e.message));
  }, []);

  return (
    <div className="page">
      <header className="page-head">
        <div>
          <p className="kicker">Ikigai · positioning · pillars</p>
          <h1>Brand Blueprint</h1>
          <p className="lede">Save writes SQLite (source of truth) and pulses the persist path on the architecture map.</p>
        </div>
        <button
          className="primary"
          onClick={() => {
            const next: Blueprint = {
              ...bp,
              pillars: pillars.split("\n").map((s) => s.trim()).filter(Boolean),
              topics: topics.split("\n").map((s) => s.trim()).filter(Boolean),
            };
            api
              .saveBrand(next)
              .then((b) => {
                setBp(b);
                setSaved(true);
                onPulse();
                setTimeout(() => setSaved(false), 1600);
              })
              .catch((e: Error) => setErr(e.message));
          }}
        >
          {saved ? "Saved" : "Save blueprint"}
        </button>
      </header>
      {err && <p className="err">{err}</p>}
      <div className="ikigai">
        <Field label="Love" value={bp.love} onChange={(v) => setBp({ ...bp, love: v })} />
        <Field label="Good at" value={bp.goodAt} onChange={(v) => setBp({ ...bp, goodAt: v })} />
        <Field label="World needs" value={bp.worldNeeds} onChange={(v) => setBp({ ...bp, worldNeeds: v })} />
        <Field label="Paid for" value={bp.paidFor} onChange={(v) => setBp({ ...bp, paidFor: v })} />
      </div>
      <div className="split">
        <label className="panel grow">
          Positioning
          <textarea value={bp.positioning} onChange={(e) => setBp({ ...bp, positioning: e.target.value })} />
        </label>
        <label className="panel grow">
          Voice
          <textarea value={bp.voice} onChange={(e) => setBp({ ...bp, voice: e.target.value })} />
        </label>
      </div>
      <div className="split">
        <label className="panel grow">
          Pillars (one per line)
          <textarea value={pillars} onChange={(e) => setPillars(e.target.value)} />
        </label>
        <label className="panel grow">
          Topics (one per line)
          <textarea value={topics} onChange={(e) => setTopics(e.target.value)} />
        </label>
      </div>
    </div>
  );
}

function Field({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <label className="panel">
      {label}
      <textarea value={value} onChange={(e) => onChange(e.target.value)} rows={4} />
    </label>
  );
}
