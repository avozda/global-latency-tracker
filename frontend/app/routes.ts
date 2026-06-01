import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("api/metrics", "routes/api.metrics.ts"),
] satisfies RouteConfig;
