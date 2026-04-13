import Section from "./section";
import { CheckCircle2 } from "lucide-react";

export default function HowItWorks() {
  return (
    <Section className="flex-col gap-10 max-w-4xl mx-auto py-16 sm:py-24">
      <div className="flex flex-col gap-4">
        <h2 className="text-3xl sm:text-4xl font-bold tracking-tight">How It Works</h2>
      </div>
      <div className="flex flex-col gap-10">
        {[
          {
            title: "Define Entrypoints",
            desc: "Tell Prune where your app starts (e.g., src/index.ts, app/page.tsx)."
          },
          {
            title: "Graph Traversal",
            desc: "Prune parses your files into ASTs and traces every import, export, and call-site."
          },
          {
            title: "Reachability Analysis",
            desc: "Anything not connected to your entrypoints is flagged."
          },
          {
            title: "Actionable Results",
            desc: "Get a clear list of what can be safely removed or reviewed."
          }
        ].map((item, i) => (
          <div key={i} className="flex flex-col sm:flex-row gap-4 sm:gap-8 items-start">
            <div className="flex items-center justify-center min-w-10 min-h-10 sm:min-w-12 sm:min-h-12 rounded-full border border-border bg-muted text-foreground font-semibold text-lg sm:text-xl">
              {i + 1}
            </div>
            <div className="flex flex-col gap-2 pt-1">
              <h3 className="font-semibold text-xl sm:text-2xl">{item.title}</h3>
              <p className="text-base sm:text-lg text-muted-foreground mt-1 sm:mt-0">{item.desc}</p>
            </div>
          </div>
        ))}
      </div>
      <div className="pt-8 border-t border-border mt-6 flex flex-col sm:flex-row items-center sm:items-start gap-4 text-center sm:text-left">
        <CheckCircle2 className="text-muted-foreground shrink-0 hidden sm:block" />
        <p className="text-lg sm:text-xl font-medium tracking-tight">Run it locally or in CI — same behavior, same results.</p>
      </div>
    </Section>
  );
}
