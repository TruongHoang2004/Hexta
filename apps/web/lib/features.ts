// Build-time feature switches (NEXT_PUBLIC_* values are inlined when the web image is built).

/** Google sign-in buttons on /login and /register. Hidden unless the build sets NEXT_PUBLIC_GOOGLE_LOGIN=true. */
export const googleLoginEnabled = process.env.NEXT_PUBLIC_GOOGLE_LOGIN === "true";
