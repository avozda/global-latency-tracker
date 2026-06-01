import { fetchMetrics } from "../lib/api";

/**
 * Same-origin proxy for the Go hub metrics endpoint. Runs server-side only so
 * the secret API key stays out of the browser; React Query fetches from here.
 */
export async function loader({ request }: { request: Request }) {
  const url = new URL(request.url);
  const limitParam = url.searchParams.get("limit");
  const limit = limitParam ? Number(limitParam) : 100;

  const records = await fetchMetrics(Number.isFinite(limit) ? limit : 100);

  return new Response(JSON.stringify(records), {
    headers: { "Content-Type": "application/json" },
  });
}
