import Section from "./section";

export default function WhyItsDifferent() {
  return (
    <Section className="flex-col gap-6 max-w-4xl mx-auto py-12 md:py-16 border-y border-border bg-muted/30">
      <h2 className="text-4xl font-bold tracking-tight">Built for speed, not just compliance</h2>
      <p className="text-xl text-muted-foreground leading-relaxed">
        Most tools either slow you down or give you results you can't trust. Prune does neither.
      </p>
      <ul className="flex flex-col gap-6 mt-4 text-lg">
        <li className="flex items-start gap-3">
          <div className="mt-1 w-2 h-2 rounded-full bg-primary shrink-0" />
          <p className="text-muted-foreground"><strong className="text-foreground font-semibold">Faster than TSC:</strong> Focused on structural reachability without full type-checking overhead.</p>
        </li>
        <li className="flex items-start gap-3">
          <div className="mt-1 w-2 h-2 rounded-full bg-primary shrink-0" />
          <p className="text-muted-foreground"><strong className="text-foreground font-semibold">Deterministic:</strong> No black-box heuristics—Prune follows your code exactly.</p>
        </li>
        <li className="flex items-start gap-3">
          <div className="mt-1 w-2 h-2 rounded-full bg-primary shrink-0" />
          <p className="text-muted-foreground"><strong className="text-foreground font-semibold">Non-Intrusive:</strong> Works alongside your existing tools without requiring project changes.</p>
        </li>
      </ul>
    </Section>
  );
}
