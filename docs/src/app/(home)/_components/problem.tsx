import Section from "./section";

export default function Problem() {
  return (
    <Section className="flex-col gap-10 max-w-4xl mx-auto py-16 sm:py-24">
      <div className="flex flex-col gap-4 text-center sm:text-left">
        <h2 className="text-3xl sm:text-4xl font-bold tracking-tight leading-tight">Your codebase is growing, but is it actually doing anything?</h2>
        <p className="text-lg sm:text-xl text-muted-foreground leading-relaxed max-w-3xl">
          As projects evolve, they accumulate technical debt in the form of "zombie" code. Files are moved, features are deprecated, and exports are left hanging. Standard linters catch unused variables, but they miss the bigger picture:
        </p>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-8 sm:gap-12">
        <div className="flex flex-col gap-3">
          <h3 className="font-semibold text-xl sm:text-2xl">Orphaned Files</h3>
          <p className="text-muted-foreground leading-relaxed text-sm sm:text-base">Entire components and modules that are never imported but still bloat your bundle and slow down your build.</p>
        </div>
        <div className="flex flex-col gap-3">
          <h3 className="font-semibold text-xl sm:text-2xl">Dead Exports</h3>
          <p className="text-muted-foreground leading-relaxed text-sm sm:text-base">Library functions and components that are exported "just in case" but have zero consumers.</p>
        </div>
        <div className="flex flex-col gap-3">
          <h3 className="font-semibold text-xl sm:text-2xl">Shadow Dependencies</h3>
          <p className="text-muted-foreground leading-relaxed text-sm sm:text-base">Code referenced in tests or documentation but completely unreachable from your production entrypoints.</p>
        </div>
        <div className="flex flex-col gap-3">
          <h3 className="font-semibold text-xl sm:text-2xl">Analysis Paralysis</h3>
          <p className="text-muted-foreground leading-relaxed text-sm sm:text-base">Static analysis tools are often too slow to run on every commit or too noisy to trust.</p>
        </div>
      </div>
      <div className="pt-4 border-t border-border mt-4">
        <p className="text-2xl font-medium tracking-tight">You’re shipping code that no one—and nothing—uses.</p>
      </div>
    </Section>
  );
}
