import { randomBytes } from "node:crypto";

import { NextRequest, NextResponse } from "next/server";

import { ADMIN_AUTH_UPSTREAM_ALLOWLIST } from "@/lib/admin-auth";

const allowedPaths = ADMIN_AUTH_UPSTREAM_ALLOWLIST;
const apiURL = process.env.CAPA_API_URL ?? "http://localhost:8080";

type RouteContext = { params: Promise<{ path: string[] }> };

export async function GET(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  if (path.join("/") === "csrf") {
    const csrfToken = randomBytes(32).toString("hex");
    const response = NextResponse.json({ csrfToken });
    response.cookies.set("capa_admin_csrf", csrfToken, {
      httpOnly: false,
      sameSite: "lax",
      secure: process.env.NODE_ENV === "production",
      path: "/",
      maxAge: 60 * 60 * 8,
    });
    return response;
  }
  return forward(request, path, "GET");
}
export async function POST(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  return forward(request, path, "POST");
}

async function forward(request: NextRequest, path: string[], method: "GET" | "POST") {
  const targetPath = path.join("/");
  if (!allowedPaths.has(targetPath)) {
    return NextResponse.json({ code: "not_found", message: "Kaynak bulunamadÄ±." }, { status: 404 });
  }

  const target = new URL(`/v1/admin/${targetPath}`, apiURL);
  const headers = new Headers();
  for (const name of ["cookie", "content-type", "origin", "x-csrf-token"]) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }

  const upstream = await fetch(target, {
    method,
    headers,
    body: method === "POST" ? await request.arrayBuffer() : undefined,
    cache: "no-store",
  });
  const responseHeaders = new Headers();
  for (const name of ["content-type", "set-cookie"]) {
    const value = upstream.headers.get(name);
    if (value) responseHeaders.set(name, value);
  }
  return new NextResponse(upstream.body, { status: upstream.status, headers: responseHeaders });
}
