import { codeToHtml } from 'shiki';
import { cn } from "@/lib/utils";

interface CodeBlockProps {
  code: string;
  lang?: string;
  filename?: string;
  className?: string;
}

export default async function CodeBlock({ code, lang = "typescript", filename, className }: CodeBlockProps) {
  const html = await codeToHtml(code, {
    lang,
    themes: {
      light: 'catppuccin-mocha'
    }
  });

  return (
    <div className={cn("rounded-4xl border border-border bg-background overflow-hidden shadow-sm", className)}>
      <div className="flex items-center px-4 py-3 border-b border-border/10 bg-background">
        <div className="flex gap-1.5">
          <div className="w-3 h-3 rounded-full bg-muted"></div>
          <div className="w-3 h-3 rounded-full bg-muted"></div>
          <div className="w-3 h-3 rounded-full bg-muted"></div>
        </div>
        {filename && (
          <div className="flex-1 text-center text-xs text-muted-foreground mr-14 font-medium">
            {filename}
          </div>
        )}
      </div>
      <div 
        className="py-3 px-4 overflow-x-auto text-sm sm:text-base [&>pre]:bg-transparent! [&>pre]:p-0! [&>pre]:m-0! [&_code]:bg-transparent! [&_code]:p-0! [&_code]:font-mono"
        dangerouslySetInnerHTML={{ __html: html }} 
      />
    </div>
  );
}
