import type { Metadata } from 'next';
import Script from 'next/script';

const BASE_URL = 'https://useprune.dev';

export const metadata: Metadata = {
  metadataBase: new URL(BASE_URL),
  title: {
    default: 'Prune — Dead Code Analyzer for JavaScript & TypeScript',
    template: '%s | Prune',
  },
  description:
    'Prune is a static analysis CLI tool that finds unreachable code, orphaned files, and unused exports in JavaScript and TypeScript projects. Powered by Tree-sitter. Built for CI/CD.',
  keywords: [
    'dead code',
    'static analysis',
    'JavaScript',
    'TypeScript',
    'unused exports',
    'orphaned files',
    'tree-sitter',
    'CLI',
    'CI/CD',
    'code quality',
    'linter',
    'code pruning',
  ],
  authors: [{ name: 'Carlos Eduardo', url: 'https://github.com/carlosedujs' }],
  creator: 'Carlos Eduardo',
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: BASE_URL,
    siteName: 'Prune',
    title: 'Prune — Dead Code Analyzer for JavaScript & TypeScript',
    description:
      'Find and remove unreachable code, orphaned files, and unused exports in JS/TS projects. Works on any OS. Fits right into your CI.',
    images: [
      {
        url: '/og/home/image.webp',
        width: 1200,
        height: 630,
        alt: 'Prune — Dead Code Analyzer for JS/TS',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Prune — Dead Code Analyzer for JavaScript & TypeScript',
    description:
      'Find and remove unreachable code, orphaned files, and unused exports in JS/TS projects.',
    images: ['/og/home/image.webp'],
    creator: '@carlosedujs',
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  alternates: {
    canonical: BASE_URL,
  },
  verification: {
    google: 'GOOGLE_SEARCH_CONSOLE_TOKEN',
  },
};

const jsonLd = {
  '@context': 'https://schema.org',
  '@type': 'SoftwareApplication',
  name: 'Prune',
  description:
    'A static analysis CLI tool that identifies dead code — unreachable files, unused exports, undeclared functions and variables — in JavaScript and TypeScript projects using Tree-sitter AST parsing.',
  applicationCategory: 'DeveloperApplication',
  operatingSystem: 'Linux, macOS, Windows',
  url: BASE_URL,
  softwareVersion: '0.0.3',
  license: 'https://opensource.org/licenses/MIT',
  author: {
    '@type': 'Person',
    name: 'Carlos Eduardo',
    url: 'https://github.com/carlosedujs',
  },
  codeRepository: 'https://github.com/carlosedujs/prune',
  programmingLanguage: 'Go',
  keywords: 'dead code, static analysis, JavaScript, TypeScript, CLI, tree-sitter',
  offers: {
    '@type': 'Offer',
    price: '0',
    priceCurrency: 'USD',
  },
};

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <Script
        id="json-ld-software"
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      {children}
    </>
  );
}

