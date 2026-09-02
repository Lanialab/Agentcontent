import { useEffect, useMemo, useRef, useState } from "react";
import type { Pulse } from "../api";
import { edges, kindColor, nodeMap, nodes, pathD, VIEW, type Kind } from "./model";

type Props = {
  pulses: Pulse[];
  compact?: boolean;
};

export function ArchitectureMap({ pulses, compact }: Props) {
  const svgRef = useRef<SVGSVGElement>(null);
  const [paused, setPaused] = useState(false);
  const [showAll, setShowAll] = useState(false);
  const [tick, setTick] = useState(0);
  const byId = useMemo(() => nodeMap(), []);

  const hot = useMemo(() => {
    const latest = pulses[0];
    const nodesHot = new Set<string>();
    const edgesHot = new Set<string>();
    const pathsHot = new Set<string>();
    if (showAll) {
      for (const e of edges) edgesHot.add(e.id);
      for (const n of nodes) nodesHot.add(n.id);
      return { nodesHot, edgesHot, pathsHot, latest };
    }
    if (latest) {
      pathsHot.add(latest.path);
      for (const id of latest.nodes ?? []) nodesHot.add(id);
      for (const e of edges) {
        if (e.paths.includes(latest.path)) edgesHot.add(e.id);
      }
    }
    return { nodesHot, edgesHot, pathsHot, latest };
  }, [pulses, showAll]);

  useEffect(() => {
    const svg = svgRef.current;
    if (!svg) return;
    if (paused) svg.pauseAnimations();
    else svg.unpauseAnimations();
  }, [paused, tick]);

  function restart() {
    setPaused(false);
    setShowAll(false);
    const svg = svgRef.current;
    if (svg) {
      svg.unpauseAnimations();
      svg.setCurrentTime(0);
    }
    setTick((t) => t + 1);
  }

  return (
    <div className={`arch ${compact ? "compact" : ""}`}>
      {!compact && (
        <div className="arch-toolbar">
          <div className="arch-controls">
            <button type="button" onClick={() => setPaused((p) => !p)}>
              {paused ? "Tiếp tục" : "Tạm dừng"}
            </button>
            <button type="button" onClick={restart}>
              Chạy lại
            </button>
            <button type="button" className={showAll ? "on" : ""} onClick={() => setShowAll((s) => !s)}>
              Hiện tất cả
            </button>
          </div>
          <div className="legend">
            <span className="lg command">command / query</span>
            <span className="lg async">async job / progress</span>
            <span className="lg persist">local persistence</span>
            <span className="lg external">external I/O</span>
          </div>
          <div className="authoritative">Business logic is authoritative. Code and technology cannot override it.</div>
        </div>
      )}
      <div className="arch-canvas">
        <svg
          ref={svgRef}
          viewBox={`0 0 ${VIEW.w} ${VIEW.h}`}
          className={paused ? "paused" : ""}
          role="img"
          aria-label="AgentContent hexagonal architecture live view"
        >
          <defs>
            <filter id="glow" x="-40%" y="-40%" width="180%" height="180%">
              <feGaussianBlur stdDeviation="2.4" result="b" />
              <feMerge>
                <feMergeNode in="b" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
            <filter id="hotglow" x="-50%" y="-50%" width="200%" height="200%">
              <feGaussianBlur stdDeviation="4.5" result="b" />
              <feMerge>
                <feMergeNode in="b" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
            <linearGradient id="coreFill" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stopColor="#102038" />
              <stop offset="100%" stopColor="#0a1424" />
            </linearGradient>
            {edges.map((ed) => {
              const a = byId[ed.from];
              const b = byId[ed.to];
              if (!a || !b) return null;
              return <path key={`def-${ed.id}`} id={`wire-${ed.id}`} d={pathD(a, b)} fill="none" />;
            })}
          </defs>

          <text x="48" y="22" className="layer-label">
            PERSONAL UI CLIENT — outside Go process
          </text>
          <rect x="32" y="28" width="1080" height="92" rx="14" className="band ui-band" />

          <polygon
            points="24,148 1108,148 1136,200 1136,848 1108,900 24,900 0,848 0,200"
            className="core-hex"
          />
          <text x="40" y="162" className="layer-label core-title">
            GO APPLICATION BACKEND — hexagonal modular monolith — process boundary
          </text>

          <LayerLabel x={40} y={158} n={1} text="Driving adapters" />
          <LayerLabel x={40} y={258} n={2} text="Inbound ports" />
          <LayerLabel x={40} y={368} n={3} text="Application use cases" />
          <LayerLabel x={40} y={498} n={4} text="Domain models — pure business rules" />
          <LayerLabel x={40} y={638} n={5} text="Outbound ports" />
          <LayerLabel x={40} y={748} n={6} text="Driven adapters" />

          <text x="1148" y="150" className="layer-label">
            EXTERNAL DRIVERS + EXTERNAL SYSTEMS
          </text>

          {edges.map((ed) => {
            const a = byId[ed.from];
            const b = byId[ed.to];
            if (!a || !b) return null;
            const isHot = hot.edgesHot.has(ed.id);
            return (
              <g key={ed.id} className={`wire kind-${ed.kind} ${isHot ? "hot" : ""}`}>
                <path
                  d={pathD(a, b)}
                  fill="none"
                  stroke={kindColor[ed.kind]}
                  strokeWidth={isHot ? 2.4 : 1.35}
                  strokeDasharray={dashFor(ed.kind)}
                  filter={isHot ? "url(#hotglow)" : "url(#glow)"}
                  className="dash"
                />
                <Particle href={`#wire-${ed.id}`} color={kindColor[ed.kind]} dur={durFor(ed.kind)} begin="0s" hot={isHot} />
                <Particle href={`#wire-${ed.id}`} color={kindColor[ed.kind]} dur={durFor(ed.kind)} begin={`${Number(durFor(ed.kind)) / 2}s`} hot={isHot} />
                {isHot && (
                  <Particle href={`#wire-${ed.id}`} color="#fff" dur="1.6s" begin="0.2s" hot />
                )}
              </g>
            );
          })}

          {nodes.map((nd) => {
            const isHot = hot.nodesHot.has(nd.id);
            return (
              <g key={nd.id} className={`box tone-${nd.tone ?? "port"} ${isHot ? "hot" : ""}`}>
                <rect x={nd.x} y={nd.y} width={nd.w} height={nd.h} rx="10" />
                <text x={nd.x + 12} y={nd.y + 22} className="box-title">
                  {nd.label}
                </text>
                {nd.sub &&
                  nd.sub.split("\n").map((line, i) => (
                    <text key={i} x={nd.x + 12} y={nd.y + 40 + i * 14} className="box-sub">
                      {line}
                    </text>
                  ))}
              </g>
            );
          })}

          {!compact && (
            <foreignObject x="40" y="918" width="1500" height="150">
              <div className="arch-notes">
                <article>
                  <h4>Core flow</h4>
                  <p>External driver → driving adapter → inbound port → use case → domain → outbound port → driven adapter.</p>
                </article>
                <article>
                  <h4>Locked scheduler</h4>
                  <p>CLI entry every 60 minutes. SQLite lease skips overlapping ticks. Missed runs do not pile up.</p>
                </article>
                <article>
                  <h4>Sheets boundary</h4>
                  <p>Google Sheets is optional sync/export only. No Apps Script. Go + SQLite stay authoritative.</p>
                </article>
                <article>
                  <h4>Outlier contract</h4>
                  <p>
                    viewsPerDay vs average viewsPerDay of recent same-channel videos. Default 2.5x (2–3x). Diagram *100 is visual seed exaggeration, not the product threshold.
                  </p>
                </article>
              </div>
            </foreignObject>
          )}
        </svg>
      </div>
      {!compact && hot.latest && (
        <div className="pulse-log">
          Last pulse: <code>{hot.latest.path}</code> · {hot.latest.label} · {hot.latest.kind}
        </div>
      )}
    </div>
  );
}

function LayerLabel({ x, y, n, text }: { x: number; y: number; n: number; text: string }) {
  return (
    <text x={x} y={y} className="layer-num">
      {n}. {text}
    </text>
  );
}

function Particle({
  href,
  color,
  dur,
  begin,
  hot,
}: {
  href: string;
  color: string;
  dur: string;
  begin: string;
  hot?: boolean;
}) {
  return (
    <circle r={hot ? 4.2 : 3} fill={color} filter="url(#glow)">
      <animateMotion dur={dur} begin={begin} repeatCount="indefinite" rotate="auto">
        <mpath href={href} />
      </animateMotion>
    </circle>
  );
}

function dashFor(kind: Kind): string {
  switch (kind) {
    case "command":
      return "7 9";
    case "async":
      return "2 8";
    case "persist":
      return "10 6";
    case "external":
      return "5 7";
    default: {
      const _n: never = kind;
      return _n;
    }
  }
}

function durFor(kind: Kind): string {
  switch (kind) {
    case "command":
      return "3.2s";
    case "async":
      return "2.4s";
    case "persist":
      return "3.8s";
    case "external":
      return "2.8s";
    default: {
      const _n: never = kind;
      return _n;
    }
  }
}
