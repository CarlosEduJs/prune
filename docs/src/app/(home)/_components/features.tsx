import Section from "./section";

export default function Features() {
  return (
    <Section className="flex-col gap-10 max-w-4xl mx-auto py-16 sm:py-24">
      <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-center sm:text-left">Key Features</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-x-12 gap-y-10">
        <div className="flex flex-col gap-3">
          <h3 className="font-semibold text-xl">Runs Anywhere</h3>
          <p className="text-muted-foreground leading-relaxed text-sm sm:text-base text-justify sm:text-left">Cross-platform support (Linux, macOS, Windows) with a single binary.</p>
        </div>
        <div className="flex flex-col gap-3">
          <h3 className="font-semibold text-xl">CI/CD Ready</h3>
          <p className="text-muted-foreground leading-relaxed text-sm sm:text-base text-justify sm:text-left">Drop it into your pipeline and fail builds when dead code is detected.</p>
        </div>
        <div className="flex flex-col gap-3">
          <h3 className="font-semibold text-xl">Confidence Scoring</h3>
          <p className="text-muted-foreground leading-relaxed text-sm sm:text-base text-justify sm:text-left">Distinguishes between "safe to delete" and "review required" cases.</p>
        </div>
        <div className="flex flex-col gap-3">
          <h3 className="font-semibold text-xl">Tree-sitter Powered</h3>
          <p className="text-muted-foreground leading-relaxed text-sm sm:text-base text-justify sm:text-left">High-performance parsing without the overhead of a full compiler.</p>
        </div>
        <div className="flex flex-col gap-3">
          <h3 className="font-semibold text-xl">Real-time Streaming</h3>
          <p className="text-muted-foreground leading-relaxed text-sm sm:text-base text-justify sm:text-left">Use <code className="bg-muted px-1.5 py-0.5 rounded border border-border text-[10px] sm:text-xs">--stream</code> to see findings as they are discovered.</p>
        </div>
        <div className="flex flex-col gap-3">
          <h3 className="font-semibold text-xl">Machine Readable</h3>
          <p className="text-muted-foreground leading-relaxed text-sm sm:text-base text-justify sm:text-left">Output in <code className="bg-muted px-1.5 py-0.5 rounded border border-border text-[10px] sm:text-xs">json</code> or <code className="bg-muted px-1.5 py-0.5 rounded border border-border text-[10px] sm:text-xs">ndjson</code> for integrations and custom tooling.</p>
        </div>
        <div className="flex flex-col gap-3">
          <h3 className="font-semibold text-xl">Path Aliases</h3>
          <p className="text-muted-foreground leading-relaxed text-sm sm:text-base text-justify sm:text-left">Resolves TypeScript path aliases like <code className="bg-muted px-1.5 py-0.5 rounded border-border text-[10px] sm:text-xs">@/</code> and <code className="bg-muted px-1.5 py-0.5 rounded border border-border text-[10px] sm:text-xs">~/</code> configured via <code className="bg-muted px-1.5 py-0.5 rounded border border-border text-[10px] sm:text-xs">ts_config</code>.</p>
        </div>
      </div>
    </Section>
  );
}
