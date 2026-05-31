import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import type { ProbeRecord } from "../lib/api";

interface LatencyChartProps {
  region: string;
  series: ProbeRecord[];
}

interface ChartPoint {
  time: string;
  ttfb_ms: number;
  total_roundtrip_ms: number;
}

function toChartData(series: ProbeRecord[]): ChartPoint[] {
  return series.map((record) => ({
    time: new Date(record.measured_at).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    }),
    ttfb_ms: Number(record.ttfb_ms.toFixed(1)),
    total_roundtrip_ms: Number(record.total_roundtrip_ms.toFixed(1)),
  }));
}

export function LatencyChart({ region, series }: LatencyChartProps) {
  const data = toChartData(series);

  return (
    <div className="rounded-2xl border border-gray-800 bg-gray-900/40 p-5">
      <div className="mb-4 flex items-center justify-between">
        <div>
          <h2 className="text-sm font-semibold text-gray-100">
            Latency over time
          </h2>
          <p className="text-xs text-gray-500">
            Region <span className="text-gray-300">{region}</span> &middot;{" "}
            {series.length} samples
          </p>
        </div>
      </div>

      {data.length === 0 ? (
        <div className="flex h-72 items-center justify-center text-sm text-gray-500">
          No samples yet for this region.
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={320}>
          <LineChart
            data={data}
            margin={{ top: 8, right: 16, bottom: 8, left: 0 }}
          >
            <CartesianGrid strokeDasharray="3 3" stroke="#1f2937" />
            <XAxis
              dataKey="time"
              stroke="#6b7280"
              tick={{ fontSize: 12 }}
              tickMargin={8}
            />
            <YAxis
              stroke="#6b7280"
              tick={{ fontSize: 12 }}
              width={48}
              unit="ms"
            />
            <Tooltip
              contentStyle={{
                backgroundColor: "#0b0f19",
                border: "1px solid #1f2937",
                borderRadius: "0.75rem",
                color: "#e5e7eb",
              }}
              labelStyle={{ color: "#9ca3af" }}
              formatter={(value) => `${value} ms`}
            />
            <Legend wrapperStyle={{ fontSize: 12 }} />
            <Line
              type="monotone"
              dataKey="ttfb_ms"
              name="TTFB (ms)"
              stroke="#38bdf8"
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4 }}
            />
            <Line
              type="monotone"
              dataKey="total_roundtrip_ms"
              name="Total round-trip (ms)"
              stroke="#a78bfa"
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4 }}
            />
          </LineChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}
