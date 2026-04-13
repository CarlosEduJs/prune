"use client";

import { cn } from "@/lib/utils";

interface SectionProps {
  children: React.ReactNode;
  className?: string;
}

export default function Section({ children, className }: SectionProps) {
  return (
    <section className={cn("flex px-6 py-8 sm:px-12 sm:py-12 w-full", className)}>
      {children}
    </section>
  );
}