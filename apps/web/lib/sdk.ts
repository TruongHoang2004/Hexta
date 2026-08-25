import { EcommerceHubSDK } from "@ubi/sdk";

// Singleton SDK instance configured to point to our API Gateway
export const sdk = new EcommerceHubSDK({
  apiUrl: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
});
