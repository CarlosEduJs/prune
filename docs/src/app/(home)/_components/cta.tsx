import Section from "./section";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { Download, Star, BookOpen, Heart } from "lucide-react";

export default function CallToAction() {
  return (
    <Section className="flex-col gap-8 max-w-4xl mx-auto py-12 md:py-16 items-center text-center">
      <h2 className="text-5xl font-bold tracking-tight">Find what you can delete in 60 seconds</h2>
      <p className="text-2xl text-muted-foreground">
        Install Prune and see how much code you can remove today.
      </p>
      <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mt-4 w-full">
        <Link href="https://github.com/carlosedujs/prune/releases/latest" className="w-full sm:w-auto">
          <Button size="lg" className="w-full sm:w-auto">
            <Download className="mr-2 h-5 w-5" />
            Install for your OS
          </Button>
        </Link>
        <Link href="https://github.com/carlosedujs/prune" className="w-full sm:w-auto">
          <Button variant="outline" size="lg" className="w-full sm:w-auto">
            <Star className="mr-2 h-5 w-5" />
            Star on GitHub
          </Button>
        </Link>
        <Link href="/docs" className="w-full sm:w-auto">
          <Button variant="outline" size="lg" className="w-full sm:w-auto">
            <BookOpen className="mr-2 h-5 w-5" />
            Read the Docs
          </Button>
        </Link>
        {/* TODO: Add a link to the sponsor page */}
        <Link href="/sponsor" className="w-full sm:w-auto">
          <Button variant="outline" size="lg" className="w-full sm:w-auto">
            <Heart className="mr-2 h-5 w-5" />
            Sponsor
          </Button>
        </Link>
      </div>
      <p className="text-base text-muted-foreground mt-4 font-medium">
        Join early and help shape the fastest code-pruning engine being built.
      </p>
    </Section>
  );
}
