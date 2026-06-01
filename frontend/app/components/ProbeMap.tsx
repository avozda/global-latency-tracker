import {
  ComposableMap,
  Geographies,
  Geography,
  Marker,
} from "react-simple-maps";
import worldMap from "world-atlas/countries-110m.json";
import type { RegionGroup } from "../lib/api";
import { PROBE_LOCATIONS, TARGET_LOCATION } from "../lib/geo";

interface ProbeMapProps {
  regions: RegionGroup[];
  selectedRegion: string | null;
  onSelectRegion: (region: string) => void;
}

function markerTone(group: RegionGroup): string {
  const { latest } = group;
  if (latest.error) return "fill-red-500 stroke-red-200";
  if (latest.status_code >= 200 && latest.status_code < 300) {
    return "fill-emerald-500 stroke-emerald-200";
  }
  return "fill-amber-500 stroke-amber-200";
}

function formatLatency(value: number): string {
  return `${value.toFixed(0)} ms`;
}

export function ProbeMap({
  regions,
  selectedRegion,
  onSelectRegion,
}: ProbeMapProps) {
  const plottedRegions = regions
    .map((group) => ({
      group,
      location: PROBE_LOCATIONS[group.region],
    }))
    .filter(
      (
        entry,
      ): entry is {
        group: RegionGroup;
        location: NonNullable<(typeof entry)["location"]>;
      } => Boolean(entry.location),
    );

  return (
    <section className="rounded-2xl border border-gray-800 bg-gray-900/40 p-5">
      <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="text-sm font-semibold text-gray-100">
            Probe network
          </h2>
          <p className="text-xs text-gray-500">
            Deployed probes measuring latency to{" "}
            <span className="text-gray-300">{TARGET_LOCATION.label}</span>.
          </p>
        </div>
        <div className="flex flex-wrap gap-3 text-xs text-gray-500">
          <span className="inline-flex items-center gap-1.5">
            <span className="h-2.5 w-2.5 rounded-full bg-emerald-500" />
            Probe
          </span>
          <span className="inline-flex items-center gap-1.5">
            <span className="h-2.5 w-2.5 rotate-45 rounded-sm bg-blue-400" />
            Monitored URL
          </span>
        </div>
      </div>

      <div className="overflow-hidden rounded-xl border border-gray-800 bg-gray-950/60">
        <ComposableMap
          projectionConfig={{ scale: 150 }}
          className="h-auto w-full"
          aria-label="World map showing probe regions and the ETH Zurich target"
        >
          <Geographies geography={worldMap}>
            {({ geographies }) =>
              geographies.map((geo) => (
                <Geography
                  key={geo.rsmKey}
                  geography={geo}
                  fill="#111827"
                  stroke="#374151"
                  strokeWidth={0.4}
                  style={{
                    default: { outline: "none" },
                    hover: { fill: "#111827", outline: "none" },
                    pressed: { outline: "none" },
                  }}
                />
              ))
            }
          </Geographies>

          {plottedRegions.map(({ group, location }) => {
            const selected = group.region === selectedRegion;
            return (
              <Marker key={group.region} coordinates={location.coordinates}>
                <g
                  role="button"
                  tabIndex={0}
                  aria-label={`${group.region} probe in ${location.city}, ${formatLatency(
                    group.latest.total_roundtrip_ms,
                  )}`}
                  className="cursor-pointer focus:outline-none"
                  onClick={() => onSelectRegion(group.region)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault();
                      onSelectRegion(group.region);
                    }
                  }}
                >
                  <title>
                    {location.city} ({group.region}) -{" "}
                    {formatLatency(group.latest.total_roundtrip_ms)}
                  </title>
                  {selected && (
                    <circle
                      r={11}
                      className="fill-blue-500/20 stroke-blue-400"
                      strokeWidth={1.5}
                    />
                  )}
                  <circle
                    r={selected ? 5.5 : 4.5}
                    className={markerTone(group)}
                    strokeWidth={1.5}
                  />
                  <text
                    textAnchor="middle"
                    y={-12}
                    className="pointer-events-none fill-gray-200 text-[10px] font-semibold"
                  >
                    {group.region}
                  </text>
                </g>
              </Marker>
            );
          })}

          <Marker coordinates={TARGET_LOCATION.coordinates}>
            <g>
              <title>
                {TARGET_LOCATION.label} - {TARGET_LOCATION.city}
              </title>
              <rect
                x={-4.5}
                y={-4.5}
                width={9}
                height={9}
                rx={1.5}
                className="origin-center rotate-45 fill-blue-400 stroke-blue-100"
                strokeWidth={1.5}
              />
              <text
                textAnchor="middle"
                y={-12}
                className="pointer-events-none fill-blue-100 text-[10px] font-semibold"
              >
                ETH Zurich
              </text>
            </g>
          </Marker>
        </ComposableMap>
      </div>
    </section>
  );
}
