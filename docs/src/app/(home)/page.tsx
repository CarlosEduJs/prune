import Link from "next/link";
import { Button } from "@/components/ui/button";
import Hero from "./_components/hero";

export default function HomePage() {
  return (
    <div className="flex flex-col">
      <header className="flex items-center justify-between px-12 py-6">
        <h1 className="text-2xl font-bold">prune</h1>
        <nav>
          <ul className="flex gap-4 items-center">
            <li><Link className="text-sm font-light" href="/">Home</Link></li>
            <li><Link className="text-sm font-light" href="/docs">Docs</Link></li>
          </ul>
        </nav>
        <Button variant={"outline"} size={"lg"}>
          Get Started
        </Button>
      </header>
      <Hero />
    </div>
  );
}
