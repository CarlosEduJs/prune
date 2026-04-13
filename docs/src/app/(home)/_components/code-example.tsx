import Section from "./section";
import { Terminal } from "lucide-react";
import CodeBlock from "../../../components/code-block";

export default function CodeExample() {
  return (
    <Section className="flex-col gap-8 max-w-4xl mx-auto py-16 sm:py-24 border-y border-border bg-muted/30">
      <div className="flex flex-col gap-4">
        <h2 className="text-3xl sm:text-4xl font-bold tracking-tight">Simple configuration. Powerful results.</h2>
        
        <p className="text-base sm:text-lg text-muted-foreground mt-8 flex items-center gap-2">
          <Terminal size={18} className="shrink-0" /> Initialize your project with a single command:
        </p>
        <CodeBlock code="prune init" lang="bash" />
        
        <p className="text-base sm:text-lg text-muted-foreground mt-8">
          Your <code className="bg-background px-1.5 py-0.5 rounded border border-border text-xs sm:text-sm">prune.yaml</code> keeps things explicit:
        </p>
        <CodeBlock 
          filename="prune.yaml" 
          lang="yaml"
          code={`project:
  name: prune-docs
  language: js-ts

scan:
  paths: [src]
  include: ["**/*.ts", "**/*.tsx"]

entrypoints:
  files:
    - src/main.ts
    - src/api/server.ts

rules:
  unused_export: enabled
  unused_function: enabled
  orphaned_file: enabled`}
        />
        
        <p className="text-base sm:text-lg text-muted-foreground mt-8 flex items-center gap-2">
          <Terminal size={18} className="shrink-0" /> Run the scan and clear the noise:
        </p>
        <CodeBlock code="prune scan --fail-on-findings" lang="bash" />
      </div>
    </Section>
  );
}
