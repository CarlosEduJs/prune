"use client";

import { useState, useEffect } from "react";
import { cn } from "@/lib/utils";

const PHRASES = [
  "Delete what doesn’t matter.",
  "Ship less. Deliver more.",
  "Cut the dead weight.",
  "Your codebase, but lighter.",
  "Not all code deserves to live.",
  "Prune your codebase.",
  "If it’s not used, why is it there?",
];

export default function Footer() {
  const [index, setIndex] = useState(0);
  const [fade, setFade] = useState(true);

  useEffect(() => {
    const interval = setInterval(() => {
      setFade(false);
      setTimeout(() => {
        setIndex((current) => (current + 1) % PHRASES.length);
        setFade(true);
      }, 500);
    }, 2000);

    return () => clearInterval(interval);
  }, []);

  return (
    <footer className="bg-foreground w-full text-background flex flex-col md:flex-row items-center md:items-baseline justify-between p-12 rounded-t-xl mt-12 sm:mt-24">
      <h1 className="text-8xl sm:text-9xl font-bold">prune</h1>
      <div className="h-8 mt-6 md:mt-0 flex items-center">
        <p 
          className={cn(
            "text-xl font-bold text-muted text-center md:text-right transition-opacity duration-500",
            fade ? "opacity-100" : "opacity-0"
          )}
        >
          {PHRASES[index]}
        </p>
      </div>
    </footer>
  );
}