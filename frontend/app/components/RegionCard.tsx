import type { ProbeRecord } from "../lib/api";

interface RegionCardProps {
  region: string;
  latest: ProbeRecord;
  selected: boolean;
  onSelect: (region: string) => void;
}

function statusTone(record: ProbeRecord): {
  label: string;
  dot: string;
  badge: string;
} {
  if (record.error) {
    return {
      label: "Error",
      dot: "bg-red-500",
      badge: "bg-red-500/10 text-red-400 ring-red-500/30",
    };
  }
  if (record.status_code >= 200 && record.status_code < 300) {
    return {
      label: String(record.status_code),
      dot: "bg-emerald-500",
      badge: "bg-emerald-500/10 text-emerald-400 ring-emerald-500/30",
    };
  }
  return {
    label: String(record.status_code || "—"),
    dot: "bg-amber-500",
    badge: "bg-amber-500/10 text-amber-400 ring-amber-500/30",
  };
}

function formatTimestamp(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

export function RegionCard({
  region,
  latest,
  selected,
  onSelect,
}: RegionCardProps) {
  const tone = statusTone(latest);

  return (
    <button
      type="button"
      onClick={() => onSelect(region)}
      aria-pressed={selected}
      className={`group w-full rounded-2xl border p-5 text-left transition focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 ${
        selected
          ? "border-blue-500 bg-blue-500/5 shadow-lg shadow-blue-500/10"
          : "border-gray-800 bg-gray-900/40 hover:border-gray-700 hover:bg-gray-900/70"
      }`}
    >
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <span className={`h-2.5 w-2.5 rounded-full ${tone.dot}`} />
          <span className="text-sm font-semibold tracking-wide text-gray-100">
            {region}
          </span>
        </div>
        <span
          className={`rounded-full px-2.5 py-0.5 text-xs font-semibold ring-1 ring-inset ${tone.badge}`}
        >
          {tone.label}
        </span>
      </div>

      <div className="mt-4 flex items-baseline gap-1">
        <span className="text-3xl font-bold text-white tabular-nums">
          {latest.total_roundtrip_ms.toFixed(0)}
        </span>
        <span className="text-sm text-gray-400">ms total</span>
      </div>

      <dl className="mt-3 grid grid-cols-2 gap-x-4 gap-y-1 text-xs text-gray-400">
        <dt>TTFB</dt>
        <dd className="text-right tabular-nums text-gray-200">
          {latest.ttfb_ms.toFixed(0)} ms
        </dd>
        <dt>Updated</dt>
        <dd className="text-right tabular-nums text-gray-200">
          {formatTimestamp(latest.measured_at)}
        </dd>
      </dl>

      <p className="mt-3 truncate text-xs text-gray-500" title={latest.target_url}>
        {latest.target_url}
      </p>
    </button>
  );
}
