import { useEffect, useState } from "react";
import { ArchitectureMap } from "./architecture/ArchitectureMap";
import { api, pageLabel, type Meta, type Page, type Pulse } from "./api";
import { AssistantPage } from "./pages/Assistant";
import { BrandPage } from "./pages/Brand";
import { CreatorsPage } from "./pages/Creators";
import { FeedPage } from "./pages/Feed";
import { ScriptsPage } from "./pages/Scripts";

const nav: Page[] = ["arch", "creators", "feed", "ai", "brand", "script"];

export default function App() {
  const [page, setPage] = useState<Page>("arch");
  const [meta, setMeta] = useState<Meta | null>(null);
  const [pulses, setPulses] = useState<Pulse[]>([]);

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
        {renderPage(page, pulses, bump)}
      </main>
    </div>
  );
}

function renderPage(page: Page, pulses: Pulse[], bump: () => void) {
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
      return <FeedPage />;
    case "ai":
      return <AssistantPage onPulse={bump} />;
    case "brand":
      return <BrandPage onPulse={bump} />;
    case "script":
      return <ScriptsPage onPulse={bump} />;
    default: {
      const _n: never = page;
      return _n;
    }
  }
}
