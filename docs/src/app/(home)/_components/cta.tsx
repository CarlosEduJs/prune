import Section from "./section";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { Download, Star, BookOpen, Heart } from "lucide-react";

export default function CallToAction() {
  return (
    <Section className="flex-col gap-10 max-w-4xl mx-auto py-20 sm:py-32 items-center text-center">
      <h2 className="text-4xl sm:text-5xl font-bold tracking-tight leading-tight px-4 sm:px-0">Find what you can delete in 60 seconds</h2>
      <p className="text-xl sm:text-2xl text-muted-foreground max-w-2xl px-4 sm:px-0">
        Install Prune and see how much code you can remove today.
      </p>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-8 w-full max-w-xl px-6 sm:px-0">
        <Link href="https://github.com/carlosedujs/prune/releases/latest" className="w-full">
          <Button size="lg" className="w-full h-12 text-base">
            <Download className="mr-2 h-5 w-5" />
            Install for your OS
          </Button>
        </Link>
        <Link href="https://github.com/carlosedujs/prune" className="w-full">
          <Button variant="outline" size="lg" className="w-full h-12 text-base">
            <Star className="mr-2 h-5 w-5" />
            Star on GitHub
          </Button>
        </Link>
        <Link href="/docs" className="w-full">
          <Button variant="outline" size="lg" className="w-full h-12 text-base">
            <BookOpen className="mr-2 h-5 w-5" />
            Read the Docs
          </Button>
        </Link>
        <Link href="https://github.com/sponsors/CarlosEduJs" className="w-full">
          <Button variant="outline" size="lg" className="w-full h-12 text-base">
            <Heart className="mr-2 h-5 w-5" />
            Sponsor
          </Button>
        </Link>
      </div>
      <p className="text-sm sm:text-base text-muted-foreground mt-8 font-medium px-4 sm:px-0 italic opacity-80">
        Join early and help shape the fastest code-pruning engine being built.
      </p>
    </Section>
  );
}
