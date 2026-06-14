import { auth } from "@clerk/nextjs/server";
import * as Sentry from "@sentry/nextjs";

type RequestInitWithDuplex = RequestInit & {
  duplex?: "half"
}

async function webrtcProxy(
  req: Request,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { getToken } = await auth();
  const token = await getToken();

  const pathParts = (await params).path;
  const path = pathParts.join("/");
  const { searchParams } = new URL(req.url);

  const API_URL = process.env.API_URL;

  if (!API_URL) {
    Sentry.logger.error("API_URL is not configured", {
      area: "webrtc-proxy",
      path,
    });
    return Response.json(
      { error: "Missing PIPECAT_API_URL" },
      { status: 500 }
    );
  }

  const query = searchParams.toString();
  const targetUrl = `${API_URL}/${path}${query ? `?${query}` : ""}`;

  const forwardedHeaders = new Headers();
  const contentType = req.headers.get("content-type");
  const accept = req.headers.get("accept");

  if (contentType) forwardedHeaders.set("content-type", contentType);
  if (accept) forwardedHeaders.set("accept", accept);
  if (token) forwardedHeaders.set("Authorization", `Bearer ${token}`);

  try {
    const fetchOptions: RequestInitWithDuplex = {
      method: req.method,
      headers: forwardedHeaders,
      body:
        req.method !== "GET" && req.method !== "HEAD"
          ? req.body
          : undefined,
      duplex: "half",
      cache: "no-store",
    }

    const upstreamResponse = await fetch(targetUrl, fetchOptions)

    const isStartRoute = path.endsWith("start");

    if (isStartRoute) {
      const json = await upstreamResponse.json();

      if (!upstreamResponse.ok) {
        const message = json?.error?.message ?? "Failed to start session";

        Sentry.logger.error("Upstream /start returned error", {
          area: "webrtc-proxy",
          path,
          status: upstreamResponse.status,
          message,
        });

        Sentry.captureException(new Error(message), {
          tags: {
            area: "webrtc-proxy",
            type: "upstream-start-error",
            status: String(upstreamResponse.status),
          },
          extra: {
            path,
            targetUrl,
            status: upstreamResponse.status,
            responseBody: json,
          },
        });

        return Response.json({ error: message }, {
          status: upstreamResponse.status,
        });
      }

      const unwrapped = json?.data ?? json;
      return Response.json(unwrapped, {
        status: upstreamResponse.status,
      });
    }

    // /sessions/:sessionId/api/offer: stream as-is, no unwrapping
    if (!upstreamResponse.ok) {
      Sentry.logger.error("Upstream /offer returned error", {
        area: "webrtc-proxy",
        path,
        status: upstreamResponse.status,
      });

      Sentry.captureException(new Error("Upstream offer error"), {
        tags: {
          area: "webrtc-proxy",
          type: "upstream-offer-error",
          status: String(upstreamResponse.status),
        },
        extra: {
          path,
          targetUrl,
          status: upstreamResponse.status,
        },
      });
    }

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
    Sentry.logger.error("WebRTC proxy crashed", {
      area: "webrtc-proxy",
      path,
    });

    Sentry.captureException(error, {
      tags: {
        area: "webrtc-proxy",
        type: "proxy-crash",
      },
      extra: {
        path,
        targetUrl,
        method: req.method,
      },
    });

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