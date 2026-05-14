import { auth } from "@clerk/nextjs/server";

async function webrtcProxy(
  req: Request,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { getToken } = await auth();
  const token = await getToken();

  const pathParts = (await params).path;
  const path = pathParts.join("/");
  const { searchParams } = new URL(req.url);

  const PIPECAT_API_URL = process.env.API_URL;

  if (!PIPECAT_API_URL) {
    return Response.json(
      { error: "Missing PIPECAT_API_URL" },
      { status: 500 }
    );
  }

  const query = searchParams.toString();
  const targetUrl = `${PIPECAT_API_URL}/${path}${query ? `?${query}` : ""}`;

  const forwardedHeaders = new Headers();
  const contentType = req.headers.get("content-type");
  const accept = req.headers.get("accept");

  if (contentType) forwardedHeaders.set("content-type", contentType);
  if (accept) forwardedHeaders.set("accept", accept);
  if (token) forwardedHeaders.set("Authorization", `Bearer ${token}`);

  try {
    const upstreamResponse = await fetch(targetUrl, {
      method: req.method,
      headers: forwardedHeaders,
      body:
        req.method !== "GET" && req.method !== "HEAD"
          ? req.body
          : undefined,
      ...({ duplex: "half" } as any),
      cache: "no-store",
    });

    const isStartRoute = path.endsWith("start");

    if (isStartRoute) {
      // ✅ /start: unwrap { data: { sessionId, iceConfig } } -> { sessionId, iceConfig }
      const json = await upstreamResponse.json();

      // unwrap if your Go backend wraps in { data: ... }
      const unwrapped = json?.data ?? json;

      return Response.json(unwrapped, {
        status: upstreamResponse.status,
      });
    }

    // ✅ /sessions/:sessionId/api/offer: stream as-is, no unwrapping
    const responseHeaders = new Headers();
    const upstreamContentType = upstreamResponse.headers.get("content-type");
    if (upstreamContentType) {
      responseHeaders.set("content-type", upstreamContentType);
    }

    return new Response(upstreamResponse.body, {
      status: upstreamResponse.status,
      statusText: upstreamResponse.statusText,
      headers: responseHeaders,
    });

  } catch (error) {
    console.error("[WebRTC proxy] Error:", error);
    return Response.json(
      { error: "Failed to proxy request to Pipecat API" },
      { status: 502 }
    );
  }
}

export const GET = webrtcProxy;
export const POST = webrtcProxy;
export const PATCH = webrtcProxy;
export const DELETE = webrtcProxy;