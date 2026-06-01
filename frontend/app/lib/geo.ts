export interface GeoPoint {
  label: string;
  city: string;
  coordinates: [longitude: number, latitude: number];
}

export const PROBE_LOCATIONS: Record<string, GeoPoint> = {
  "US-West": {
    label: "US-West",
    city: "Los Angeles / LAX",
    coordinates: [-118.41, 33.94],
  },
  "South America": {
    label: "South America",
    city: "Sao Paulo / GRU",
    coordinates: [-46.47, -23.44],
  },
  Europe: {
    label: "Europe",
    city: "Amsterdam / AMS",
    coordinates: [4.77, 52.31],
  },
  Asia: {
    label: "Asia",
    city: "Singapore / SIN",
    coordinates: [103.99, 1.35],
  },
  Africa: {
    label: "Africa",
    city: "Johannesburg / JNB",
    coordinates: [28.24, -26.14],
  },
};

export const TARGET_LOCATION: GeoPoint = {
  label: "ETH Zurich",
  city: "Zurich",
  coordinates: [8.5417, 47.3769],
};
