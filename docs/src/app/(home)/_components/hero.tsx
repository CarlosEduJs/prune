import Section from "./section";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { Download } from 'lucide-react';
import { FaWindows, FaLinux, FaApple } from "react-icons/fa";
import { SiReadthedocs } from "react-icons/si";

export default function Hero() {
  return (
    <Section className="py-20 min-h-[80vh] flex-col justify-center items-center w-full mx-auto">
      <div className="flex flex-col max-w-2xl gap-8 text-center sm:text-left items-center sm:items-start">
        <h1 className="text-4xl sm:text-6xl font-bold leading-tight">Zero out the dead weight in your codebase</h1>
        <p className="text-lg sm:text-xl font-light leading-relaxed">Find and remove unreachable code, orphaned files, and unused exports in JS/TS projects, all this in just a few seconds. Works on any OS. Fits right into your CI.</p>
        <div className="flex flex-col sm:flex-row items-center gap-6 w-full">
          <div className="flex flex-col sm:flex-row items-center gap-4 w-full sm:w-auto">
            <Link href="/docs" className="w-full sm:w-auto">
              <Button size={"lg"} className="w-full sm:w-auto">
                <SiReadthedocs />
                Get Started
              </Button>
            </Link>
            <Link href="https://github.com/carlosedujs/prune/releases/latest" className="w-full sm:w-auto">
              <Button variant={"outline"} size={"lg"} className="w-full sm:w-auto">
                <Download />
                Install Prune
              </Button>
            </Link>
          </div>
          <div className="flex items-center gap-3 py-2 px-4 rounded-full bg-accent">
            <h2 className="text-xs font-light text-muted-foreground">Available on</h2>
            <div className="flex items-center gap-3">
              <FaWindows className="text-lg" />
              <FaLinux className="text-lg" />
              <FaApple className="text-lg" />
            </div>
          </div>
        </div>
      </div>
    </Section>
  );
}