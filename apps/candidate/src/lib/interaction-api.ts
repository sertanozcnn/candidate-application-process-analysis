export type InteractionEvent = {
  event_id: string;
  sequence: number;
  elapsed_ms: number;
  type: string;
  field_code?: string;
  metadata: Record<string, unknown>;
};

type SessionResponse = { id: string; position_code: string };

export class InteractionRequestError extends Error {
  constructor(public readonly status: number) {
    super("Interaction request failed");
    this.name = "InteractionRequestError";
  }
}

export async function createInteractionSession(positionCode: string): Promise<string> {
  const apiURL = process.env.NEXT_PUBLIC_CAPA_API_URL ?? "http://localhost:8080";
  const response = await fetch(new URL("/v1/candidate/interaction-sessions", apiURL), {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ position_code: positionCode }),
  });
  if (!response.ok) throw new InteractionRequestError(response.status);
  const body = (await response.json()) as SessionResponse;
  return body.id;
}

export async function sendInteractionEvents(sessionID: string, events: InteractionEvent[]): Promise<void> {
  const apiURL = process.env.NEXT_PUBLIC_CAPA_API_URL ?? "http://localhost:8080";
  const response = await fetch(new URL(`/v1/candidate/interaction-sessions/${encodeURIComponent(sessionID)}/events`, apiURL), {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ events }),
    keepalive: true,
  });
  if (!response.ok) throw new InteractionRequestError(response.status);
}
