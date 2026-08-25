import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  const token = request.cookies.get("auth_token")?.value;
  const isAuthPage = request.nextUrl.pathname.startsWith("/login") || request.nextUrl.pathname.startsWith("/register");
  
  // Example protected route paths
  const isProtectedRoute = request.nextUrl.pathname.startsWith("/tenant") || request.nextUrl.pathname.startsWith("/dashboard");

  if (!token && isProtectedRoute) {
    // Redirect to login if accessing protected route without token
    return NextResponse.redirect(new URL("/login", request.url));
  }

  if (token && isAuthPage) {
    // Redirect to home or dashboard if accessing auth pages while already logged in
    // Change this to your default authenticated route
    return NextResponse.redirect(new URL("/tenant", request.url));
  }

  return NextResponse.next();
}

// See "Matching Paths" below to learn more
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico, sitemap.xml, robots.txt (metadata files)
     */
    '/((?!api|_next/static|_next/image|favicon.ico|sitemap.xml|robots.txt).*)',
  ],
};
