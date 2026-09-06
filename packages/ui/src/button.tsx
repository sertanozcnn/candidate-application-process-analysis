"use client";

import { motion, useReducedMotion, type HTMLMotionProps } from "motion/react";
import { FiArrowRight, FiArrowUpRight } from "react-icons/fi";

type Appearance = {
  children: string;
  variant?: "primary" | "secondary";
  arrow?: "right" | "diagonal" | false;
  className?: string;
};

const variants = {
  primary: "bg-linear-to-r from-capa-brand via-capa-violet to-capa-brand text-capa-panel shadow-lg shadow-capa-brand/20 inset-ring inset-ring-capa-brand/15 hover:shadow-capa-brand/30",
  secondary: "bg-linear-to-r from-capa-panel via-capa-surface to-capa-panel text-capa-ink shadow-sm inset-ring inset-ring-capa-line hover:shadow-md",
};

function classes(variant: "primary" | "secondary", extra = "") {
  return [
    "inline-flex h-12 shrink-0 items-center justify-center gap-3 rounded-full border-0 bg-origin-border bg-size-[200%_100%] bg-position-[0%_50%] px-6 text-sm font-semibold transition-[background-position,box-shadow] duration-500 ease-out hover:bg-position-[100%_50%] focus-visible:bg-position-[100%_50%] focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-capa-violet/30 motion-reduce:transition-none",
    variants[variant], extra,
  ].join(" ");
}

function ButtonContent({ children, arrow = "right" }: Appearance) {
  const reducedMotion = useReducedMotion();
  const Icon = arrow === "diagonal" ? FiArrowUpRight : FiArrowRight;
  return (
    <>
      <span className="relative block overflow-hidden leading-6">
        <span className="sr-only">{children}</span>
        <motion.span aria-hidden="true" className="block"
          variants={{ rest: { y: "0%" }, hover: { y: reducedMotion ? "0%" : "-100%" } }}
          transition={{ duration: 0.24, ease: "easeOut" }}>{children}</motion.span>
        <motion.span aria-hidden="true" className="absolute inset-0 block"
          variants={{ rest: { y: "100%" }, hover: { y: reducedMotion ? "100%" : "0%" } }}
          transition={{ duration: 0.24, ease: "easeOut" }}>{children}</motion.span>
      </span>
      {arrow && <Icon aria-hidden="true" className="size-[18px] shrink-0" />}
    </>
  );
}

export function Button({ children, variant = "primary", arrow = "right", className = "", type = "button", disabled, ...props }: Omit<HTMLMotionProps<"button">, "children"> & Appearance) {
  return (
    <motion.button {...props} data-capa-button type={type} disabled={disabled} className={classes(variant, className)}
      onClick={(event) => {
        if (props["aria-disabled"] === true || props["aria-disabled"] === "true") {
          event.preventDefault();
          return;
        }
        props.onClick?.(event);
      }}
      initial="rest" whileHover="hover" whileFocus="hover">
      <ButtonContent arrow={arrow}>{children}</ButtonContent>
    </motion.button>
  );
}

export function ButtonLink({ children, variant = "primary", arrow = "right", className = "", ...props }: Omit<HTMLMotionProps<"a">, "children"> & Appearance & { href: string }) {
  return (
    <motion.a {...props} data-capa-button className={classes(variant, className)} initial="rest" whileHover="hover" whileFocus="hover">
      <ButtonContent arrow={arrow}>{children}</ButtonContent>
    </motion.a>
  );
}
