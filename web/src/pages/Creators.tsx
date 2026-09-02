import { useEffect, useState } from "react";
import { api, type Channel, type Group } from "../api";

export function CreatorsPage({ onPulse }: { onPulse: () => void }) {
  const [groups, setGroups] = useState<Group[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [handle, setHandle] = useState("");
  const [groupId, setGroupId] = useState("");
  const [newGroup, setNewGroup] = useState("");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState("");

  async function load() {
    const [g, c] = await Promise.all([api.groups(), api.channels(true)]);
    setGroups(g);
    setChannels(c);
    if (!groupId && g[0]) setGroupId(g[0].id);
  }

  useEffect(() => {
    load().catch((e: Error) => setErr(e.message));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function wrap(label: string, fn: () => Promise<void>) {
    setErr("");
    setBusy(label);
    try {
      await fn();
      onPulse();
      await load();
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setBusy("");
    }
  }

  return (
    <div className="page">
      <header className="page-head">
        <div>
          <p className="kicker">Research domain</p>
          <h1>Creators & Groups</h1>
          <p className="lede">Add, validate, edit, hide, move. Scan pulls recent videos through VideoSourcePort — stub if no YouTube key.</p>
        </div>
        <button className="primary" disabled={!!busy} onClick={() => wrap("scan", async () => { await api.scan(); })}>
          {busy === "scan" ? "Scanning…" : "Scan now"}
        </button>
      </header>
      {err && <p className="err">{err}</p>}
      <div className="row">
        <form
          className="panel grow"
          onSubmit={(e) => {
            e.preventDefault();
            wrap("add", async () => {
              await api.addChannel({ groupId, handle });
              setHandle("");
            });
          }}
        >
          <h3>Add creator</h3>
          <label>
            Handle / URL
            <input value={handle} onChange={(e) => setHandle(e.target.value)} placeholder="f8official" />
          </label>
          <label>
            Group
            <select value={groupId} onChange={(e) => setGroupId(e.target.value)}>
              {groups.map((g) => (
                <option key={g.id} value={g.id}>
                  {g.name}
                </option>
              ))}
            </select>
          </label>
          <button type="submit" className="primary" disabled={!handle || !!busy}>
            Add
          </button>
        </form>
        <form
          className="panel"
          onSubmit={(e) => {
            e.preventDefault();
            wrap("group", async () => {
              await api.createGroup(newGroup);
              setNewGroup("");
            });
          }}
        >
          <h3>New group</h3>
          <input value={newGroup} onChange={(e) => setNewGroup(e.target.value)} placeholder="Collab" />
          <button type="submit" disabled={!newGroup.trim()}>
            Create
          </button>
        </form>
      </div>
      <div className="group-board">
        {groups.map((g) => (
          <section key={g.id} className="group-col">
            <h3>{g.name}</h3>
            {channels
              .filter((c) => c.groupId === g.id)
              .map((c) => (
                <article key={c.id} className={`card ${c.hidden ? "dim" : ""}`}>
                  <div className="card-top">
                    <strong>{c.title}</strong>
                    <span className="handle">@{c.handle}</span>
                  </div>
                  <p className="muted">{c.description || "—"}</p>
                  <div className="pills">
                    <span className={c.validated ? "ok" : "warn"}>{c.validated ? "validated" : "unvalidated"}</span>
                    {c.hidden && <span className="warn">hidden</span>}
                  </div>
                  <div className="actions">
                    <button onClick={() => wrap("val", async () => { await api.validateChannel(c.id); })}>Validate</button>
                    <button onClick={() => wrap("hide", async () => { await api.hideChannel(c.id, !c.hidden); })}>
                      {c.hidden ? "Unhide" : "Hide"}
                    </button>
                    <select
                      value={c.groupId}
                      onChange={(e) => wrap("move", async () => { await api.moveChannel(c.id, e.target.value); })}
                    >
                      {groups.map((og) => (
                        <option key={og.id} value={og.id}>
                          Move → {og.name}
                        </option>
                      ))}
                    </select>
                  </div>
                </article>
              ))}
          </section>
        ))}
      </div>
    </div>
  );
}
