import { EcommerceHubSDK } from "@ecommercehub/sdk";

// Helper to get token (e.g. from cookies or localStorage)
const getToken = () => {
  if (typeof window !== "undefined") {
    return localStorage.getItem("auth_token");
  }
  return null;
};

// Singleton SDK instance configured to point to our API Gateway
export const sdk = new EcommerceHubSDK({
  apiUrl: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
  getToken,
});
