import { useEffect, useRef, useState } from "react";
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
  const [chatBusy, setChatBusy] = useState(false);
  const [ideaBusy, setIdeaBusy] = useState(false);
  const [pendingAssistant, setPendingAssistant] = useState(false);
  const transcriptRef = useRef<HTMLDivElement>(null);

  async function load() {
    const [c, i, t] = await Promise.all([api.chat(), api.ideas(), api.templates()]);
    setChat(c);
    setIdeas(i);
    setTemplates(t);
  }

  useEffect(() => {
    load().catch((e: Error) => setErr(e.message));
  }, []);

  useEffect(() => {
    const el = transcriptRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [chat, pendingAssistant]);

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
          <div className="transcript" ref={transcriptRef}>
            {chat.map((m) => (
              <div key={m.id} className={`bubble ${m.role}`}>
                <span>{m.role}</span>
                <pre>{m.content}</pre>
              </div>
            ))}
            {pendingAssistant && (
              <div className="bubble assistant pending">
                <span>assistant · đang nghĩ…</span>
                <div className="typing-dots" aria-hidden="true">
                  <i />
                  <i />
                  <i />
                </div>
              </div>
            )}
          </div>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              const trimmed = message.trim();
              if (!trimmed || chatBusy) return;
              const optimisticId = `local-${Date.now()}`;
              const optimistic: ChatMessage = {
                id: optimisticId,
                role: "user",
                content: trimmed,
                createdAt: new Date().toISOString(),
              };
              setChat((prev) => [...prev, optimistic]);
              setMessage("");
              setErr("");
              setChatBusy(true);
              setPendingAssistant(true);
              api
                .sendChat({ message: trimmed, templateId, mentions: mentions.split(/[\s,]+/).filter(Boolean) })
                .then(() => {
                  onPulse();
                  return load();
                })
                .catch((er: Error) => {
                  setErr(er.message);
                  setChat((prev) => prev.filter((m) => m.id !== optimisticId));
                  setMessage(trimmed);
                })
                .finally(() => {
                  setChatBusy(false);
                  setPendingAssistant(false);
                });
            }}
          >
            <textarea value={message} onChange={(e) => setMessage(e.target.value)} placeholder="Hỏi Intelligence…" />
            <button className="primary" disabled={chatBusy || !message.trim()}>
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
            disabled={ideaBusy}
            onClick={() => {
              setIdeaBusy(true);
              setErr("");
              api
                .generateIdea({ prompt, templateId, mentions: mentions.split(/[\s,]+/).filter(Boolean) })
                .then(() => {
                  onPulse();
                  return load();
                })
                .catch((er: Error) => setErr(er.message))
                .finally(() => setIdeaBusy(false));
            }}
          >
            Generate idea
          </button>
          {ideaBusy && <p className="muted-note">Đang generate…</p>}
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
