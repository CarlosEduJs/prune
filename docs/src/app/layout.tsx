import type { Metadata } from 'next';
import { RootProvider } from 'fumadocs-ui/provider/next';
import './global.css';
import { Inter, Sora, Public_Sans, Fira_Code } from 'next/font/google';
import { cn } from "@/lib/utils";

const firaCodeFiraCode = Fira_Code({subsets:['cyrillic','cyrillic-ext','greek','greek-ext','latin','latin-ext','symbols2'],weight:['300','400','500','600','700'],variable:'--font-fira-code'});

const publicSansPublicSans = Public_Sans({subsets:['latin','latin-ext','vietnamese'],weight:['100','200','300','400','500','600','700','800','900'],variable:'--font-public-sans'});

const soraSora = Sora({subsets:['latin','latin-ext'],weight:['100','200','300','400','500','600','700','800'],variable:'--font-sora'});

const inter = Inter({
  subsets: ['latin'],
});

export const metadata: Metadata = {
  metadataBase: new URL('https://useprune.dev'),
  title: {
    default: 'Prune — Dead Code Analyzer for JavaScript & TypeScript',
    template: '%s | Prune',
  },
  description:
    'Prune is a static analysis CLI tool that finds unreachable code, orphaned files, and unused exports in JavaScript and TypeScript projects.',
  // Add your verification tokens below when you have them:
  // verification: {
  //   google: 'YOUR_GOOGLE_VERIFICATION_TOKEN',
  //   yandex: 'YOUR_YANDEX_TOKEN',
  // },
};

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={cn(inter.className, soraSora.variable, publicSansPublicSans.variable, firaCodeFiraCode.variable)} suppressHydrationWarning>
      <body className="flex flex-col min-h-screen">
        <RootProvider>{children}</RootProvider>
      </body>
    </html>
  );
}

