import { useEffect, useState } from "react";
import { api, type ChatMessage, type Idea, type Template } from "../api";

export function AssistantPage({ onPulse }: { onPulse: () => void }) {
  const [chat, setChat] = useState<ChatMessage[]>([]);
  const [ideas, setIdeas] = useState<Idea[]>([]);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [message, setMessage] = useState("");
  const [prompt, setPrompt] = useState("Một idea khớp pillar Research & outliers");
  const [templateId, setTemplateId] = useState("outlier-remix");
  const [mentions, setMentions] = useState("@f8official");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  async function load() {
    const [c, i, t] = await Promise.all([api.chat(), api.ideas(), api.templates()]);
    setChat(c);
    setIdeas(i);
    setTemplates(t);
  }

  useEffect(() => {
    load().catch((e: Error) => setErr(e.message));
  }, []);

  return (
    <div className="page">
      <header className="page-head">
        <div>
          <p className="kicker">Intelligence domain</p>
          <h1>AI Assistant & Ideas</h1>
          <p className="lede">Chat, templates, mentions. Generate idea pulses the AI path on the live architecture map.</p>
        </div>
      </header>
      {err && <p className="err">{err}</p>}
      <div className="split">
        <section className="panel chat">
          <h3>Chat</h3>
          <div className="transcript">
            {chat.map((m) => (
              <div key={m.id} className={`bubble ${m.role}`}>
                <span>{m.role}</span>
                <pre>{m.content}</pre>
              </div>
            ))}
          </div>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              setBusy(true);
              api
                .sendChat({ message, templateId, mentions: mentions.split(/[\s,]+/).filter(Boolean) })
                .then(() => {
                  setMessage("");
                  onPulse();
                  return load();
                })
                .catch((er: Error) => setErr(er.message))
                .finally(() => setBusy(false));
            }}
          >
            <textarea value={message} onChange={(e) => setMessage(e.target.value)} placeholder="Hỏi Intelligence…" />
            <button className="primary" disabled={busy || !message.trim()}>
              Send
            </button>
          </form>
        </section>
        <section className="panel">
          <h3>Generate idea</h3>
          <label>
            Template
            <select value={templateId} onChange={(e) => setTemplateId(e.target.value)}>
              {templates.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Mentions
            <input value={mentions} onChange={(e) => setMentions(e.target.value)} />
          </label>
          <textarea value={prompt} onChange={(e) => setPrompt(e.target.value)} />
          <button
            className="primary"
            disabled={busy}
            onClick={() => {
              setBusy(true);
              api
                .generateIdea({ prompt, templateId, mentions: mentions.split(/[\s,]+/).filter(Boolean) })
                .then(() => {
                  onPulse();
                  return load();
                })
                .catch((er: Error) => setErr(er.message))
                .finally(() => setBusy(false));
            }}
          >
            Generate idea
          </button>
          <ul className="idea-list">
            {ideas.map((idea) => (
              <li key={idea.id}>
                <strong>{idea.title}</strong>
                <p>{idea.body.slice(0, 220)}</p>
              </li>
            ))}
          </ul>
        </section>
      </div>
    </div>
  );
}
