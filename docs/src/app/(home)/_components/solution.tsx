import Section from "./section";

export default function Solution() {
  return (
    <Section className="flex-col gap-6 max-w-4xl mx-auto py-16 sm:py-24 border-y border-border bg-muted/30">
      <h2 className="text-3xl sm:text-4xl font-bold tracking-tight leading-tight">Precision-engineered code pruning</h2>
      <p className="text-lg sm:text-xl text-muted-foreground leading-relaxed">
        Prune builds a complete dependency graph of your application starting from your actual entrypoints—including resolving TypeScript path aliases like <code className="bg-muted px-1.5 py-0.5 rounded border border-border text-[10px] sm:text-xs">@/</code> and <code className="bg-muted px-1.5 py-0.5 rounded border border-border text-[10px] sm:text-xs">~/</code>. If a file, function, or export isn't reachable from the root, it's flagged as dead code.
      </p>
      <p className="text-lg sm:text-xl text-muted-foreground leading-relaxed">
        Built in Go and powered by Tree-sitter, Prune is fast enough to run as a pre-commit hook or as a blocking gate in your CI/CD pipeline. No more manual hunting—just a clean, leaner codebase.
      </p>
    </Section>
  );
}
