import { useEffect, useMemo, useRef, useState } from "react";
import type { Pulse } from "../api";
import {
  driveToExt,
  extSystems,
  HEX,
  hexBandPolygon,
  hexPointsAttr,
  hexWires,
  kindColor,
  layoutHex,
  uiModules,
  uiToDrive,
  type Chip,
  type Kind,
  type LaidChip,
} from "./hex";

type Props = {
  pulses: Pulse[];
  compact?: boolean;
};

export function ArchitectureMap({ pulses, compact }: Props) {
  const svgRef = useRef<SVGSVGElement>(null);
  const viewRef = useRef<HTMLDivElement>(null);
  const scaleRef = useRef(1);
  const [paused, setPaused] = useState(false);
  const [showAll, setShowAll] = useState(false);
  const [tick, setTick] = useState(0);
  const [scale, setScale] = useState(1);

  const laid = useMemo(() => layoutHex(), []);
  const wires = useMemo(() => hexWires(laid), [laid]);

  const hot = useMemo(() => {
    const latest = pulses[0];
    const ids = new Set<string>();
    const wireHot = new Set<string>();
    const pathHot = latest?.path ?? "";
    if (showAll) {
      for (const w of wires) wireHot.add(w.id);
      for (const c of [...uiModules, ...extSystems]) ids.add(c.id);
      for (const layer of laid) for (const c of layer.chips) ids.add(c.id);
      return { ids, wireHot, latest, pathHot };
    }
    if (latest) {
      for (const id of latest.nodes ?? []) ids.add(id);
      for (const c of uiModules) if (c.paths.includes(latest.path)) ids.add(c.id);
      for (const c of extSystems) if (c.paths.includes(latest.path)) ids.add(c.id);
      for (const layer of laid) {
        for (const c of layer.chips) {
          if (c.paths.includes(latest.path)) ids.add(c.id);
        }
      }
      for (const w of wires) {
        if (w.paths.includes(latest.path)) wireHot.add(w.id);
      }
    }
    return { ids, wireHot, latest, pathHot };
  }, [pulses, showAll, wires, laid]);

  useEffect(() => {
    scaleRef.current = scale;
  }, [scale]);

  useEffect(() => {
    const svg = svgRef.current;
    if (!svg) return;
    if (paused) svg.pauseAnimations();
    else svg.unpauseAnimations();
  }, [paused, tick]);

  useEffect(() => {
    const el = viewRef.current;
    if (!el || compact) return;
    let startDist = 0;
    let startScale = 1;
    function dist(t: TouchList): number {
      return Math.hypot(t[0].clientX - t[1].clientX, t[0].clientY - t[1].clientY);
    }
    function onStart(e: TouchEvent) {
      if (e.touches.length === 2) {
        startDist = dist(e.touches);
        startScale = scaleRef.current;
      }
    }
    function onMove(e: TouchEvent) {
      if (e.touches.length === 2 && startDist > 0) {
        e.preventDefault();
        const next = startScale * (dist(e.touches) / startDist);
        setScale(Math.min(2.6, Math.max(0.7, next)));
      }
    }
    el.addEventListener("touchstart", onStart, { passive: true });
    el.addEventListener("touchmove", onMove, { passive: false });
    return () => {
      el.removeEventListener("touchstart", onStart);
      el.removeEventListener("touchmove", onMove);
    };
  }, [compact]);

  function restart() {
    setPaused(false);
    setShowAll(false);
    setScale(1);
    const svg = svgRef.current;
    if (svg) {
      svg.unpauseAnimations();
      svg.setCurrentTime(0);
    }
    setTick((t) => t + 1);
  }

  function onWheel(e: React.WheelEvent) {
    if (compact) return;
    if (!(e.ctrlKey || e.metaKey || e.nativeEvent.ctrlKey)) return;
    e.preventDefault();
    const dir = e.deltaY > 0 ? -0.1 : 0.1;
    setScale((s) => Math.min(2.6, Math.max(0.7, s + dir)));
  }

  const hexSvg = (
    <svg
      ref={svgRef}
      viewBox={`0 0 ${HEX.w} ${HEX.h}`}
      className={paused ? "paused" : ""}
      role="img"
      aria-label="Go hexagonal modular monolith"
      style={{ transform: compact ? undefined : `scale(${scale})`, transformOrigin: "center top" }}
    >
      <defs>
        <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="2.2" result="b" />
          <feMerge>
            <feMergeNode in="b" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
        <filter id="hotglow" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="4.2" result="b" />
          <feMerge>
            <feMergeNode in="b" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
        <linearGradient id="hexFill" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#10243c" />
          <stop offset="55%" stopColor="#0b1728" />
          <stop offset="100%" stopColor="#0a1410" />
        </linearGradient>
        {wires.map((w) => (
          <path key={`def-${w.id}`} id={`wire-${w.id}`} d={w.d} fill="none" />
        ))}
      </defs>

      <polygon points={hexPointsAttr(HEX.cx, HEX.cy, HEX.r)} className="hex-shell" fill="url(#hexFill)" />
      {[0.86, 0.72, 0.56, 0.4, 0.26].map((f) => (
        <polygon
          key={f}
          points={hexPointsAttr(HEX.cx, HEX.cy, HEX.r * f)}
          className="hex-ring"
        />
      ))}

      {laid.map((layer) => (
        <polygon
          key={`band-${layer.id}`}
          points={hexBandPolygon(HEX.cx, HEX.cy, HEX.r - 3, layer.y0, layer.y1)}
          className={`hex-band band-${layer.id}`}
        />
      ))}

      {wires.map((w) => {
        const isHot = hot.wireHot.has(w.id);
        return (
          <g key={w.id} className={`wire kind-${w.kind} ${isHot ? "hot" : ""}`}>
            <path
              d={w.d}
              fill="none"
              stroke={kindColor[w.kind]}
              strokeWidth={isHot ? 2.3 : 1.25}
              strokeDasharray={dashFor(w.kind)}
              filter={isHot ? "url(#hotglow)" : "url(#glow)"}
              className="dash"
            />
            <Particle href={`#wire-${w.id}`} color={kindColor[w.kind]} dur={durFor(w.kind)} begin="0s" hot={isHot} />
            <Particle
              href={`#wire-${w.id}`}
              color={kindColor[w.kind]}
              dur={durFor(w.kind)}
              begin={`${Number.parseFloat(durFor(w.kind)) / 2}s`}
              hot={isHot}
            />
            {isHot && <Particle href={`#wire-${w.id}`} color="#fff" dur="1.5s" begin="0.15s" hot />}
          </g>
        );
      })}

      {laid.map((layer) => (
        <g key={layer.id}>
          <text x={HEX.cx} y={layer.y0 + 12} textAnchor="middle" className="layer-num">
            {layer.n}. {layer.title}
          </text>
          {layer.chips.map((c) => (
            <ChipBox key={c.id} chip={c} hot={hot.ids.has(c.id)} />
          ))}
        </g>
      ))}
    </svg>
  );

  if (compact) {
    return <div className="arch compact">{hexSvg}</div>;
  }

  return (
    <div className="arch hex-layout">
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
          <button type="button" onClick={() => setScale((s) => Math.min(2.6, s + 0.15))} aria-label="Zoom in">
            +
          </button>
          <button type="button" onClick={() => setScale((s) => Math.max(0.7, s - 0.15))} aria-label="Zoom out">
            −
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

      <div className="hex-stage">
        <section className="side-col ui-col">
          <p className="col-kicker">Personal UI Client</p>
          <p className="col-note">Outside Go process</p>
          {uiModules.map((c) => (
            <SideCard key={c.id} chip={c} hot={hot.ids.has(c.id) || showAll} />
          ))}
        </section>

        <div className="hex-core">
          <p className="hex-caption">
            <strong>GO APPLICATION BACKEND</strong>
            <span>HEXAGONAL MODULAR MONOLITH — process boundary</span>
          </p>
          <div className="hex-viewport" ref={viewRef} onWheel={onWheel}>
            {hexSvg}
          </div>
        </div>

        <section className="side-col ext-col">
          <p className="col-kicker">External systems</p>
          <p className="col-note">Drivers outside the hex</p>
          {extSystems.map((c) => (
            <SideCard key={c.id} chip={c} hot={hot.ids.has(c.id) || (showAll && c.paths.length > 0)} />
          ))}
        </section>
      </div>

      <ConnectorHint hotPath={hot.pathHot} showAll={showAll} />

      <div className="arch-notes html-notes">
        <article>
          <h4>Core flow</h4>
          <p>UI client → driving adapter → inbound port → use case → domain → outbound port → driven adapter → external I/O.</p>
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
      {hot.latest && (
        <div className="pulse-log">
          Last pulse: <code>{hot.latest.path}</code> · {hot.latest.label} · {hot.latest.kind}
        </div>
      )}
    </div>
  );
}

function SideCard({ chip, hot }: { chip: Chip; hot: boolean }) {
  return (
    <article className={`side-card tone-${chip.tone} ${hot ? "hot" : ""}`}>
      <strong>{chip.label}</strong>
      {chip.sub && <span>{chip.sub}</span>}
    </article>
  );
}

function ChipBox({ chip, hot }: { chip: LaidChip; hot: boolean }) {
  const fs = chip.w < 86 ? 8.2 : 9.4;
  return (
    <g className={`box tone-${chip.tone} ${hot ? "hot" : ""}`}>
      <rect x={chip.x} y={chip.y} width={chip.w} height={chip.h} rx="7" />
      <text x={chip.x + chip.w / 2} y={chip.y + (chip.sub ? chip.h / 2 - 4 : chip.h / 2 + 3)} textAnchor="middle" className="box-title" fontSize={fs}>
        {chip.label}
      </text>
      {chip.sub && (
        <text x={chip.x + chip.w / 2} y={chip.y + chip.h / 2 + 10} textAnchor="middle" className="box-sub" fontSize={7.2}>
          {chip.sub}
        </text>
      )}
    </g>
  );
}

function ConnectorHint({
  hotPath,
  showAll,
}: {
  hotPath: string;
  showAll: boolean;
}) {
  const left = uiToDrive.filter((x) => showAll || x.paths.includes(hotPath));
  const right = driveToExt.filter((x) => showAll || x.paths.includes(hotPath));
  if (left.length === 0 && right.length === 0) {
    return <p className="connector-hint">Pinch or use + / − to zoom the hex. Side columns stack on a phone.</p>;
  }
  return (
    <p className="connector-hint live">
      {left.length > 0 && <span className="lg command">UI → driving adapters</span>}
      {right.length > 0 && <span className="lg external">driven adapters → external I/O</span>}
    </p>
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
    <circle r={hot ? 4 : 2.7} fill={color} filter="url(#glow)">
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
      return "3.1s";
    case "async":
      return "2.3s";
    case "persist":
      return "3.6s";
    case "external":
      return "2.7s";
    default: {
      const _n: never = kind;
      return _n;
    }
  }
}
