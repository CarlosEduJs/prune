import type { MetadataRoute } from 'next';

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: '*',
        allow: '/',
        disallow: ['/api/', '/og/'],
      },
    ],
    sitemap: 'https://prune-rose.vercel.app/sitemap.xml',
    host: 'https://prune-rose.vercel.app',
  };
}
