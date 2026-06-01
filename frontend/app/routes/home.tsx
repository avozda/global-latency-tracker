import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { Route } from "./+types/home";
import {
  fetchClientMetrics,
  fetchMetrics,
  groupByRegion,
  type RegionGroup,
} from "../lib/api";
import { RegionCard } from "../components/RegionCard";
import { LatencyChart } from "../components/LatencyChart";

const REFRESH_INTERVAL_MS = 15_000;

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Global Latency Tracker" },
    {
      name: "description",
      content: "Real-time latency monitoring across global regions.",
    },
  ];
}

export async function loader() {
  try {
    const records = await fetchMetrics(100);
    return { regions: groupByRegion(records), error: null as string | null };
  } catch (error) {
    return {
      regions: [] as RegionGroup[],
      error: error instanceof Error ? error.message : "Failed to load metrics",
    };
  }
}

export default function Home({ loaderData }: Route.ComponentProps) {
  const { error: initialError } = loaderData;

  const {
    data: regions,
    isFetching,
    error: queryError,
  } = useQuery({
    queryKey: ["metrics"],
    queryFn: async () => groupByRegion(await fetchClientMetrics(100)),
    initialData: loaderData.regions,
    refetchInterval: REFRESH_INTERVAL_MS,
  });

  const error =
    queryError instanceof Error
      ? queryError.message
      : regions.length > 0
        ? null
        : initialError;

  const [selectedRegion, setSelectedRegion] = useState<string | null>(
    regions[0]?.region ?? null,
  );

  useEffect(() => {
    if (regions.length === 0) {
      if (selectedRegion !== null) setSelectedRegion(null);
      return;
    }
    if (!regions.some((group) => group.region === selectedRegion)) {
      setSelectedRegion(regions[0].region);
    }
  }, [regions, selectedRegion]);

  const selected = useMemo(
    () =>
      regions.find((group) => group.region === selectedRegion) ?? regions[0],
    [regions, selectedRegion],
  );

  return (
    <main className="min-h-screen bg-gray-950 text-gray-100">
      <div className="mx-auto max-w-6xl px-4 py-10">
        <header className="mb-8 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-white">
              Global Latency Tracker
            </h1>
            <p className="text-sm text-gray-400">
              Most recent HTTP status and latency per region.
            </p>
          </div>
          <div className="flex items-center gap-2 text-xs text-gray-500">
            <span
              className={`h-2 w-2 rounded-full ${
                isFetching
                  ? "animate-pulse bg-amber-500"
                  : "bg-emerald-500"
              }`}
            />
            {isFetching
              ? "Refreshing…"
              : `Auto-refreshing every ${REFRESH_INTERVAL_MS / 1000}s`}
          </div>
        </header>

        {error ? (
          <div className="rounded-2xl border border-red-500/30 bg-red-500/10 p-5 text-sm text-red-300">
            <p className="font-semibold">Could not reach the hub API.</p>
            <p className="mt-1 text-red-400/80">{error}</p>
            <p className="mt-2 text-xs text-red-400/60">
              Check that the Go hub is running and that{" "}
              <code className="font-mono">API_BASE_URL</code> and{" "}
              <code className="font-mono">API_KEY</code> are configured.
            </p>
          </div>
        ) : regions.length === 0 ? (
          <div className="rounded-2xl border border-gray-800 bg-gray-900/40 p-8 text-center text-sm text-gray-400">
            No probe results yet. Once a probe reports in, regions will appear
            here.
          </div>
        ) : (
          <div className="flex flex-col gap-8">
            <section>
              <h2 className="mb-3 text-xs font-semibold uppercase tracking-wider text-gray-500">
                Regions
              </h2>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {regions.map((group) => (
                  <RegionCard
                    key={group.region}
                    region={group.region}
                    latest={group.latest}
                    selected={group.region === selected?.region}
                    onSelect={setSelectedRegion}
                  />
                ))}
              </div>
            </section>

            {selected && (
              <section>
                <LatencyChart
                  region={selected.region}
                  series={selected.series}
                />
              </section>
            )}
          </div>
        )}
      </div>
    </main>
  );
}
