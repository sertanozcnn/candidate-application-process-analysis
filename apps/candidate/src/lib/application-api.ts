export type ApplicationInput = {
  full_name: string;
  email: string;
  position_code: string;
  experience: string;
};

export class ApplicationRequestError extends Error {
  constructor(public readonly status: number) {
    super("Application request failed");
    this.name = "ApplicationRequestError";
  }
}

export async function submitApplication(input: ApplicationInput, interactionSessionID?: string | null): Promise<void> {
  const apiURL = process.env.NEXT_PUBLIC_CAPA_API_URL ?? "http://localhost:8080";
  const response = await fetch(new URL("/v1/candidate/applications", apiURL), {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(interactionSessionID ? { "X-Interaction-Session-ID": interactionSessionID } : {}),
    },
    body: JSON.stringify(input),
  });

  if (response.status !== 201) {
    throw new ApplicationRequestError(response.status);
  }
}
