import { useEffect, useState } from "react";
import { ArchitectureMap } from "./architecture/ArchitectureMap";
import { api, pageLabel, type Meta, type Page, type Pulse } from "./api";
import { AssistantPage } from "./pages/Assistant";
import { BrandPage } from "./pages/Brand";
import { CreatorsPage } from "./pages/Creators";
import { FeedPage } from "./pages/Feed";
import { ScriptsPage } from "./pages/Scripts";

const nav: Page[] = ["arch", "creators", "feed", "ai", "brand", "script"];

export type DemoHandoff = {
  prompt?: string;
  mentions?: string;
  templateId?: string;
  topic?: string;
  title?: string;
  sourceVideoTitle?: string;
  sourceChannel?: string;
  ideaId?: string;
};

const demoSteps: { page: Page; label: string; n: number }[] = [
  { page: "feed", label: "Outlier", n: 1 },
  { page: "ai", label: "Idea", n: 2 },
  { page: "script", label: "Script", n: 3 },
];

function stepIndex(page: Page): number {
  const i = demoSteps.findIndex((s) => s.page === page);
  return i;
}

export default function App() {
  const [page, setPage] = useState<Page>("arch");
  const [meta, setMeta] = useState<Meta | null>(null);
  const [pulses, setPulses] = useState<Pulse[]>([]);
  const [handoff, setHandoff] = useState<DemoHandoff>({});

  useEffect(() => {
    api.meta().then(setMeta).catch(() => undefined);
    const es = new EventSource("/api/events");
    es.onmessage = (ev) => {
      try {
        const p = JSON.parse(ev.data) as Pulse;
        setPulses((cur) => [p, ...cur].slice(0, 12));
      } catch {
        /* ignore */
      }
    };
    return () => es.close();
  }, []);

  function bump() {
    /* SSE delivers the pulse; this forces Architecture to stay mounted via state in parent. */
  }

  const demoIdx = stepIndex(page);
  const showStepper = demoIdx >= 0;

  return (
    <div className="shell">
      <aside>
        <div className="brand">
          <span className="mark" />
          <div>
            <strong>AgentContent</strong>
            <em>Personal creator OS</em>
          </div>
        </div>
        <nav>
          {nav.map((p) => (
            <button key={p} className={page === p ? "active" : ""} onClick={() => setPage(p)}>
              {pageLabel(p)}
            </button>
          ))}
        </nav>
        <div className="mini">
          <ArchitectureMap pulses={pulses} compact />
        </div>
        <div className="stubs">
          <span className={meta?.youtubeStub ? "stub" : "live"}>{meta?.youtubeStub ? "YouTube stub" : "YouTube live"}</span>
          <span className={meta?.aiStub ? "stub" : "live"}>{meta?.aiStub ? "AI stub" : "AI live"}</span>
          <span className="stub">Sheets off</span>
        </div>
      </aside>
      <main>
        {showStepper && (
          <div className="demo-stepper" role="navigation" aria-label="Demo Alex">
            <div className="demo-stepper-label">Demo Alex</div>
            <div className="demo-steps">
              {demoSteps.map((s, i) => {
                const on = i === demoIdx;
                const done = i < demoIdx;
                return (
                  <button
                    key={s.page}
                    type="button"
                    className={`demo-step${on ? " on" : ""}${done ? " done" : ""}`}
                    onClick={() => {
                      if (i <= demoIdx) setPage(s.page);
                    }}
                    disabled={i > demoIdx}
                  >
                    <span className="demo-step-n">{done ? "✓" : s.n}</span>
                    <span>
                      {s.n} {s.label}
                    </span>
                  </button>
                );
              })}
            </div>
            {(handoff.sourceVideoTitle || handoff.title || handoff.topic) && (
              <p className="demo-stepper-ctx">
                {handoff.sourceVideoTitle
                  ? `Remix: ${handoff.sourceVideoTitle}${handoff.sourceChannel ? ` · @${handoff.sourceChannel}` : ""}`
                  : handoff.title
                    ? `Idea: ${handoff.title}`
                    : `Topic: ${handoff.topic}`}
              </p>
            )}
          </div>
        )}
        {renderPage(page, pulses, bump, handoff, setHandoff, setPage)}
      </main>
    </div>
  );
}

function renderPage(
  page: Page,
  pulses: Pulse[],
  bump: () => void,
  handoff: DemoHandoff,
  setHandoff: (h: DemoHandoff | ((prev: DemoHandoff) => DemoHandoff)) => void,
  setPage: (p: Page) => void,
) {
  switch (page) {
    case "arch":
      return (
        <div className="page arch-page">
          <header className="page-head">
            <div>
              <p className="kicker">Hexagonal modular monolith</p>
              <h1>Architecture Live View</h1>
              <p className="lede">
                Wires carry moving signals. Scan, generate idea, or save blueprint and the matching path flares.
              </p>
            </div>
          </header>
          <ArchitectureMap pulses={pulses} />
        </div>
      );
    case "creators":
      return <CreatorsPage onPulse={bump} />;
    case "feed":
      return (
        <FeedPage
          onRemixIdea={(next) => {
            setHandoff({
              prompt: next.prompt,
              mentions: next.mentions,
              templateId: next.templateId,
              sourceVideoTitle: next.sourceVideoTitle,
              sourceChannel: next.sourceChannel,
            });
            setPage("ai");
          }}
        />
      );
    case "ai":
      return (
        <AssistantPage
          onPulse={bump}
          handoff={handoff}
          onHandoff={(next) => setHandoff((prev) => ({ ...prev, ...next }))}
          onGoScript={(next) => {
            setHandoff((prev) => ({ ...prev, ...next }));
            setPage("script");
          }}
        />
      );
    case "brand":
      return <BrandPage onPulse={bump} />;
    case "script":
      return <ScriptsPage onPulse={bump} handoff={handoff} />;
    default: {
      const _n: never = page;
      return _n;
    }
  }
}
