import { EcommerceHubSDK } from "@ubi/sdk";

import Cookies from "js-cookie";

// Helper to get token (from cookies)
const getToken = () => {
  if (typeof window !== "undefined") {
    return Cookies.get("auth_token") || null;
  }
  return null;
};

// Singleton SDK instance configured to point to our API Gateway
export const sdk = new EcommerceHubSDK({
  apiUrl: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
  getToken,
});
