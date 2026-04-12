import defaultMdxComponents from 'fumadocs-ui/mdx';
import * as LucideIcons from 'lucide-react';
import type { MDXComponents } from 'mdx/types';

export function getMDXComponents(components?: MDXComponents) {
  return {
    ...defaultMdxComponents,
    ...(LucideIcons as unknown as MDXComponents),
    ...components,
  } satisfies MDXComponents;
}

export const useMDXComponents = getMDXComponents;

declare global {
  type MDXProvidedComponents = ReturnType<typeof getMDXComponents>;
}
