import { ImageResponse } from '@takumi-rs/image-response';

export const revalidate = false;
export const runtime = 'nodejs';

export function GET() {
  return new ImageResponse(
    (
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          width: '100%',
          height: '100%',
          backgroundColor: 'rgb(10,10,10)',
          color: 'white',
          padding: '72px',
          fontFamily: 'sans-serif',
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '10px',
            marginBottom: '40px',
          }}
        >
          <div
            style={{
              display: 'flex',
              backgroundColor: 'rgba(140,200,80,0.15)',
              borderRadius: '999px',
              padding: '6px 18px',
              fontSize: '22px',
              color: 'rgb(160,220,80)',
              fontWeight: 600,
            }}
          >
            Static Analysis · JavaScript &amp; TypeScript
          </div>
        </div>

        <span
          style={{
            fontSize: '88px',
            fontWeight: 800,
            lineHeight: 1.05,
            letterSpacing: '-2px',
            color: 'rgb(248,248,248)',
          }}
        >
          prune
        </span>

        <p
          style={{
            fontSize: '38px',
            color: 'rgba(220,220,220,0.75)',
            marginTop: '24px',
            lineHeight: 1.35,
            maxWidth: '820px',
          }}
        >
          Zero out the dead weight in your codebase.
          Find unreachable code, orphaned files, and unused exports.
        </p>

        <div
          style={{
            display: 'flex',
            flexDirection: 'row',
            gap: '16px',
            marginTop: 'auto',
          }}
        >
          {['Tree-sitter Powered', 'CI/CD Ready', 'Cross-platform', 'Open Source'].map((label) => (
            <div
              key={label}
              style={{
                display: 'flex',
                backgroundColor: 'rgba(255,255,255,0.07)',
                borderRadius: '8px',
                padding: '10px 22px',
                fontSize: '22px',
                color: 'rgba(220,220,220,0.85)',
                fontWeight: 500,
              }}
            >
              {label}
            </div>
          ))}
        </div>
      </div>
    ),
    {
      width: 1200,
      height: 630,
      format: 'webp',
    },
  );
}
