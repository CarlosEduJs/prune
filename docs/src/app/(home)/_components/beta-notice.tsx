import Section from "./section";

export default function BetaNotice() {
  return (
    <Section className="flex-col gap-6 max-w-4xl mx-auto py-12 md:py-16">
      <div className="p-8 md:p-10 border border-border rounded-2xl shadow-sm flex flex-col gap-5">
        <h2 className="text-2xl font-bold flex items-center gap-3">
          We are in Beta
        </h2>
        <p className="text-lg text-muted-foreground leading-relaxed">
          Prune is in active development and already handling real-world codebases. You may encounter edge cases with dynamic imports or unconventional patterns.
        </p>
        <p className="text-lg text-muted-foreground leading-relaxed">
          We’re looking for early adopters to help refine the engine. Found a false positive? Have an idea? We want your feedback.
        </p>
      </div>
    </Section>
  );
}
