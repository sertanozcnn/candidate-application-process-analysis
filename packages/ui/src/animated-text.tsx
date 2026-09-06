"use client";

import { motion } from "motion/react";

type AnimatedTextProps = {
  children: string;
  className?: string;
};

export function AnimatedText({ children, className = "" }: AnimatedTextProps) {
  const words = children.split(" ");

  return (
    <motion.span
      aria-label={children}
      className={["inline-flex flex-wrap", className].join(" ")}
      initial="rest"
      whileHover="hover"
    >
      {words.map((word, index) => (
        <motion.span
          aria-hidden="true"
          className="mr-[0.28em] inline-block"
          key={`${word}-${index}`}
          variants={{
            rest: { opacity: 1, y: 0 },
            hover: { opacity: 0.82, y: -2 },
          }}
          transition={{ delay: index * 0.018, duration: 0.18, ease: "easeOut" }}
        >
          {word}
        </motion.span>
      ))}
    </motion.span>
  );
}
