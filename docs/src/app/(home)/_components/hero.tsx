import Section from "./section";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { Download } from 'lucide-react';
import { FaWindows, FaLinux, FaApple } from "react-icons/fa";
import { SiReadthedocs } from "react-icons/si";
import Terminal from "@/components/terminal";

export default function Hero() {
  return (
    <Section className="py-0 h-[90vh] items-center w-fit mx-auto">
      <div className="flex flex-col max-w-2xl gap-6">
        <h1 className="text-6xl font-bold">Zero out the dead weight in your codebase</h1>
        <p className="text-xl font-light">Find and remove unreachable code, orphaned files, and unused exports in JS/TS projects, all this in just a few seconds. Works on any OS. Fits right into your CI.</p>
        <div className="flex items-center gap-4">
          <Link href="/docs">
            <Button size={"lg"}>
              <SiReadthedocs />
              Get Started
            </Button>
          </Link>
          <Link href="https://github.com/carlosedujs/prune/releases/latest">
            <Button variant={"outline"} size={"lg"} className="w-fit">
              <Download />
              Install Prune
            </Button>
          </Link>
          <div className="flex items-center gap-2">
            <h2 className="text-xs">Writed in Go, Compiled to a single binary</h2>
            <FaWindows />
            <FaLinux />
            <FaApple />
          </div>
        </div>
      </div>
    </Section>
  );
}