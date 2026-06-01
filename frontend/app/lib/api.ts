export interface ProbeRecord {
  id: number;
  target_url: string;
  region: string;
  status_code: number;
  dns_lookup_ms: number;
  tcp_connection_ms: number;
  tls_handshake_ms: number;
  server_processing_ms: number;
  ttfb_ms: number;
  total_roundtrip_ms: number;
  measured_at: string;
  error: string | null;
  created_at: string;
}

export interface RegionGroup {
  region: string;
  latest: ProbeRecord;
  series: ProbeRecord[];
}

export async function fetchMetrics(limit = 100): Promise<ProbeRecord[]> {
  const baseURL = process.env.API_BASE_URL;
  const apiKey = process.env.API_KEY;

  if (!baseURL) {
    throw new Error("API_BASE_URL environment variable is not set");
  }
  if (!apiKey) {
    throw new Error("API_KEY environment variable is not set");
  }

  const url = `${baseURL.replace(/\/$/, "")}/api/metrics?limit=${limit}`;
  const response = await fetch(url, {
    headers: { "X-API-Key": apiKey },
  });

  if (!response.ok) {
    throw new Error(
      `Hub API request failed: ${response.status} ${response.statusText}`,
    );
  }

  return (await response.json()) as ProbeRecord[];
}

export async function fetchClientMetrics(limit = 100): Promise<ProbeRecord[]> {
  const response = await fetch(`/api/metrics?limit=${limit}`);

  if (!response.ok) {
    let message = `Metrics request failed: ${response.status} ${response.statusText}`;

    try {
      const payload = (await response.json()) as { error?: unknown };
      if (typeof payload.error === "string" && payload.error.length > 0) {
        message = payload.error;
      }
    } catch {}

    throw new Error(message);
  }

  return (await response.json()) as ProbeRecord[];
}

export function groupByRegion(records: ProbeRecord[]): RegionGroup[] {
  const byRegion = new Map<string, ProbeRecord[]>();

  for (const record of records) {
    const region = record.region || "unknown";
    const existing = byRegion.get(region);
    if (existing) {
      existing.push(record);
    } else {
      byRegion.set(region, [record]);
    }
  }

  const groups: RegionGroup[] = [];
  for (const [region, regionRecords] of byRegion) {
    const sorted = [...regionRecords].sort(
      (a, b) =>
        new Date(a.measured_at).getTime() - new Date(b.measured_at).getTime(),
    );
    groups.push({
      region,
      latest: sorted[sorted.length - 1],
      series: sorted,
    });
  }

  groups.sort((a, b) => a.region.localeCompare(b.region));
  return groups;
}
