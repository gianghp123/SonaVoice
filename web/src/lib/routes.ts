export const API_ROUTES = {
  SESSIONS: {
    LIST: "/sessions",
    BY_ID: (id: string) => `/sessions/${id}`,
    CANCEL: (id: string) => `/sessions/${id}/cancel`,
    MESSAGES: (id: string) => `/sessions/${id}/messages`,
  },
  PROFILE: {
    GET: "/profile",
    UPSERT: "/profile",
    UPDATE: "/profile",
  },
  LEARNING: {
    GRAMMAR: {
      ANALYZE: "/learning/grammar/analyze",
    },
  },
} as const

export const PROXY_ROUTES = {
  WEBRTC: {
    START: (sessionId: string) => `/api/proxy/webrtc/sessions/${sessionId}/start`,
    OFFER: (sessionId: string) => `/api/proxy/webrtc/sessions/${sessionId}/api/offer`,
  },
} as const

export const AUTH_ROUTES = {
  SIGN_IN: "/sign-in",
  SIGN_UP: "/sign-up",
} as const

export const PAGE_ROUTES = {
  HOME: "/",
  SESSION: {
    DETAIL: (id: string) => `/sessions/${id}`,
  },
  CHAT: {
    SESSION: (id: string) => `/chat/${id}`,
  },
} as const
