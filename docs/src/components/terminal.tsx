"use client";

import { useState, useEffect } from "react";

interface TerminalProps {
  lines: string[];
  className?: string;
}

export default function Terminal({ lines, className }: TerminalProps) {
  const [displayedLines, setDisplayedLines] = useState<string[]>([]);
  const [currentLineIndex, setCurrentLineIndex] = useState(0);
  const [displayedText, setDisplayedText] = useState("");

  useEffect(() => {
    if (currentLineIndex >= lines.length) return;

    const currentLine = lines[currentLineIndex];
    
    if (displayedText.length < currentLine.length) {
      const timeout = setTimeout(() => {
        setDisplayedText(currentLine.slice(0, displayedText.length + 1));
      }, Math.random() * 30 + 10);
      return () => clearTimeout(timeout);
    }

    const nextTimeout = setTimeout(() => {
      setDisplayedLines((prev) => [...prev, currentLine]);
      setCurrentLineIndex((prev) => prev + 1);
      setDisplayedText("");
    }, 10);

    return () => clearTimeout(nextTimeout);
  }, [currentLineIndex, displayedText, lines]);

  return (
    <div className="rounded-lg bg-card w-full max-w-xl">
      <div className="rounded-md border border-border w-full flex flex-col">
        <div className="flex items-center gap-2 px-3 py-2 bg-muted border-b border-border shrink-0">
          <div className="w-3 h-3 rounded-full bg-destructive" />
          <div className="w-3 h-3 rounded-full bg-yellow-500" />
          <div className="w-3 h-3 rounded-full bg-green-500" />
          <span className="ml-2 text-xs text-muted-foreground">bash</span>
        </div>
        <div className="flex-1 p-4 font-mono text-xs overflow-x-auto overflow-y-auto min-h-48 max-h-80">
          {displayedLines.map((line, i) => (
            <div key={i} className="text-card-foreground whitespace-pre-wrap">
              {line}
            </div>
          ))}
          {displayedText && (
            <div className="text-card-foreground whitespace-pre-wrap">
              {displayedText}
              <span className="animate-pulse">▋</span>
            </div>
          )}
          {currentLineIndex >= lines.length && (
            <div className="text-card-foreground animate-pulse">▋</div>
          )}
        </div>
      </div>
    </div>
  );
}