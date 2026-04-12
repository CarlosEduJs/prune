"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export default function Navbar() {
  const [isScrolled, setIsScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 20);
    };
    
    // Check initial scroll position
    handleScroll();
    
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  return (
    <div className={cn(
      "fixed top-0 left-0 right-0 z-50 flex justify-center w-full pointer-events-none transition-all duration-500 ease-out",
      isScrolled ? "pt-4 sm:pt-6" : "pt-0"
    )}>
      <header className={cn(
        "flex items-center justify-between pointer-events-auto transition-all duration-500 ease-out border",
        isScrolled 
          ? "w-[95%] max-w-3xl px-6 py-3 bg-background/80 backdrop-blur-md rounded-full shadow-sm border-border" 
          : "w-full max-w-full px-12 py-6 bg-transparent border-transparent rounded-none"
      )}>
        <h1 className={cn(
          "font-bold transition-all duration-500",
          isScrolled ? "text-xl" : "text-2xl"
        )}>
          prune
        </h1>
        
        <nav>
          <ul className="flex gap-4 sm:gap-6 items-center">
            <li>
              <Link className={cn(
                "font-medium hover:text-foreground transition-all duration-500",
                isScrolled ? "text-sm text-foreground/80" : "text-sm text-foreground/60"
              )} href="/">
                Home
              </Link>
            </li>
            <li>
              <Link className={cn(
                "font-medium hover:text-foreground transition-all duration-500",
                isScrolled ? "text-sm text-foreground/80" : "text-sm text-foreground/60"
              )} href="/docs">
                Docs
              </Link>
            </li>
          </ul>
        </nav>
        
        <Link href="/docs" className="hidden sm:inline-block">
          <Button variant="outline" size={isScrolled ? "sm" : "default"} className="rounded-full transition-all duration-500">
            Get Started
          </Button>
        </Link>
      </header>
    </div>
  );
}
