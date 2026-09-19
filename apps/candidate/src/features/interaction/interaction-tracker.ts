"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import { createInteractionSession, InteractionRequestError, sendInteractionEvents, type InteractionEvent } from "@/lib/interaction-api";

const MAX_QUEUE_SIZE = 50;
const FLUSH_INTERVAL_MS = 5000;
const MAX_RETRIES = 2;

type FieldCode = "full_name" | "email" | "position_code" | "experience";
type EventType = "session_start" | "field_focus" | "field_blur" | "field_paste" | "validation_error" | "visibility_change" | "session_end";

function newEventID() {
  return typeof crypto.randomUUID === "function" ? crypto.randomUUID() : `${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

export function useInteractionTracker(positionCode: string) {
  const sessionID = useRef<string | null>(null);
  const [activeSessionID, setActiveSessionID] = useState<string | null>(null);
  const [sessionVersion, setSessionVersion] = useState(0);
  const sessionPosition = useRef("");
  const startedAt = useRef(0);
  const sequence = useRef(0);
  const queue = useRef<InteractionEvent[]>([]);
  const retryCount = useRef(0);
  const flushing = useRef(false);
  const retryTimer = useRef<number | null>(null);
  const flushRef = useRef<() => Promise<void>>(() => Promise.resolve());

  const flush = useCallback(async () => {
    if (flushing.current || !sessionID.current || queue.current.length === 0) return;
    flushing.current = true;
    const batch = queue.current.splice(0, MAX_QUEUE_SIZE);
    try {
      await sendInteractionEvents(sessionID.current, batch);
      retryCount.current = 0;
    } catch (error) {
      if (error instanceof InteractionRequestError && (error.status === 401 || error.status === 403)) {
        queue.current = [];
        sessionID.current = null;
        setActiveSessionID(null);
        setSessionVersion((current) => current + 1);
        return;
      }
      queue.current = [...batch, ...queue.current].slice(0, MAX_QUEUE_SIZE);
      retryCount.current += 1;
      if (retryCount.current <= MAX_RETRIES) {
        retryTimer.current = window.setTimeout(() => void flushRef.current(), retryCount.current * 1000);
      }
    } finally {
      flushing.current = false;
    }
  }, []);
  useEffect(() => {
    flushRef.current = flush;
  }, [flush]);

  const track = useCallback((type: EventType, fieldCode?: FieldCode, metadata: Record<string, unknown> = {}) => {
    if (!sessionID.current || queue.current.length >= MAX_QUEUE_SIZE) return;
    sequence.current += 1;
    queue.current.push({
      event_id: newEventID(),
      sequence: sequence.current,
      elapsed_ms: Math.max(0, Math.round(performance.now() - startedAt.current)),
      type,
      ...(fieldCode ? { field_code: fieldCode } : {}),
      metadata,
    });
    if (queue.current.length >= MAX_QUEUE_SIZE) void flush();
  }, [flush]);

  useEffect(() => {
    if (!positionCode) return;
    let cancelled = false;
    sessionPosition.current = positionCode;
    sessionID.current = null;
    queue.current = [];
    sequence.current = 0;
    startedAt.current = performance.now();

    void createInteractionSession(positionCode).then((id) => {
      if (cancelled) return;
      sessionID.current = id;
      setActiveSessionID(id);
      startedAt.current = performance.now();
      track("session_start");
    }).catch(() => {
      sessionID.current = null;
    });

    return () => { cancelled = true; };
  }, [positionCode, sessionVersion, track]);

  useEffect(() => {
    const interval = window.setInterval(() => void flush(), FLUSH_INTERVAL_MS);
    const onVisibilityChange = () => {
      track("visibility_change", undefined, { state: document.visibilityState === "hidden" ? "hidden" : "visible" });
      if (document.visibilityState === "hidden") void flush();
    };
    document.addEventListener("visibilitychange", onVisibilityChange);
    return () => {
      window.clearInterval(interval);
      document.removeEventListener("visibilitychange", onVisibilityChange);
      if (retryTimer.current) window.clearTimeout(retryTimer.current);
    };
  }, [flush, track]);

  const finish = useCallback(() => {
    track("session_end");
    void flush();
  }, [flush, track]);

  return { track, finish, sessionID: activeSessionID };
}
