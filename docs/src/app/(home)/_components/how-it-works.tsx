import Section from "./section";
import { CheckCircle2 } from "lucide-react";

export default function HowItWorks() {
  return (
    <Section className="flex-col gap-8 max-w-4xl mx-auto py-12 md:py-16">
      <div className="flex flex-col gap-4">
        <h2 className="text-4xl font-bold tracking-tight">How It Works</h2>
      </div>
      <div className="flex flex-col gap-6">
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
          <div key={i} className="flex gap-6 items-start">
            <div className="flex items-center justify-center min-w-10 min-h-10 rounded-full border border-border bg-muted text-foreground font-semibold text-lg">
              {i + 1}
            </div>
            <div className="flex flex-col gap-2 pt-1">
              <h3 className="font-semibold text-2xl">{item.title}</h3>
              <p className="text-lg text-muted-foreground">{item.desc}</p>
            </div>
          </div>
        ))}
      </div>
      <div className="pt-8 border-t border-border mt-4 flex items-center gap-3">
        <CheckCircle2 className="text-muted-foreground" />
        <p className="text-xl font-medium tracking-tight">Run it locally or in CI — same behavior, same results.</p>
      </div>
    </Section>
  );
}
