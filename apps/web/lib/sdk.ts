import { HextaSDK } from "@hexta/sdk";

// API origin. An empty value (production image) means same-origin requests; unset falls back to the local API.
export const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// Singleton SDK instance configured to point to our API Gateway
export const sdk = new HextaSDK({
  apiUrl,
  onAuthExpired: () => {
    if (typeof window !== "undefined" && window.location.pathname !== "/login") {
      // eslint-disable-next-line @next/next/no-location-assign-relative-destination
      window.location.href = "/login";
    }
  },
});
