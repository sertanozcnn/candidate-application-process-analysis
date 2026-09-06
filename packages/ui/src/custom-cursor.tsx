"use client";

import { motion, useMotionValue } from "motion/react";
import { useEffect, useId, useRef, useState } from "react";

export function CustomCursor() {
  const gradientId = useId();
  const cursor = useRef<HTMLDivElement>(null);
  const [overButton, setOverButton] = useState(false);
  const x = useMotionValue(-40);
  const y = useMotionValue(-40);

  useEffect(() => {
    const media = window.matchMedia("(hover: hover) and (pointer: fine)");
    const root = document.documentElement;
    const cursorClass = "[&_*]:cursor-none";
    const hide = () => {
      setOverButton(false);
      root.classList.remove(cursorClass);
      if (cursor.current) cursor.current.style.visibility = "hidden";
    };
    const move = (event: PointerEvent) => {
      if (!media.matches || event.pointerType === "touch" || !cursor.current) {
        hide();
        return;
      }
      x.set(event.clientX - 3);
      y.set(event.clientY - 3);
      cursor.current.style.visibility = "visible";
      root.classList.add(cursorClass);
      setOverButton(event.target instanceof Element && !!event.target.closest("[data-capa-button]"));
    };
    window.addEventListener("pointermove", move);
    window.addEventListener("blur", hide);
    root.addEventListener("pointerleave", hide);
    media.addEventListener("change", hide);
    return () => {
      hide();
      window.removeEventListener("pointermove", move);
      window.removeEventListener("blur", hide);
      root.removeEventListener("pointerleave", hide);
      media.removeEventListener("change", hide);
    };
  }, [x, y]);

  return (
    <motion.div ref={cursor} aria-hidden="true"
      className="pointer-events-none invisible fixed left-0 top-0 z-[9999] h-7 w-7 drop-shadow-sm"
      style={{ x, y }}>
      <span className={`absolute -left-3 -top-3 size-8 rounded-full bg-capa-rose/15 ring-1 ring-capa-brand/25 transition-opacity duration-200 motion-reduce:transition-none ${overButton ? "opacity-100" : "opacity-0"}`} />
      <svg width="28" height="28" viewBox="0 0 28 28" fill="none">
        <defs>
          <linearGradient id={gradientId} x1="4" y1="4" x2="21" y2="24" gradientUnits="userSpaceOnUse">
            <stop stopColor="var(--color-capa-blue-soft)" />
            <stop offset="0.5" stopColor="var(--color-capa-blue)" />
            <stop offset="1" stopColor="var(--color-capa-blue-strong)" />
          </linearGradient>
        </defs>
        <path d="M3 3L23 12L13 15L9 25L3 3Z" fill={`url(#${gradientId})`} stroke="var(--color-capa-panel)" strokeWidth="2" strokeLinejoin="round" />
      </svg>
    </motion.div>
  );
}
