import type { ReactNode } from 'react';
import type { ImageResponseOptions } from '@takumi-rs/image-response';

export interface GenerateProps {
    title: ReactNode;
    description?: ReactNode;
}

export async function getImageResponseOptions(): Promise<ImageResponseOptions> {
    return {
        width: 1200,
        height: 630,
        format: 'webp',
    };
}

export function generate({ title, description }: GenerateProps) {
    const siteName = 'prune';
    const primaryTextColor = 'rgb(240,240,240)';

    return (
        <div
            style={{
                display: 'flex',
                flexDirection: 'column',
                width: '100%',
                height: '100%',
                color: 'white',
                backgroundColor: 'rgb(15,15,15)',
            }}
        >
            <div
                style={{
                    display: 'flex',
                    flexDirection: 'column',
                    width: '100%',
                    height: '100%',
                    padding: '4rem',
                }}
            >
                <span
                    style={{
                        fontWeight: 600,
                        fontSize: '76px',
                    }}
                >
                    {title}
                </span>
                <p
                    style={{
                        fontSize: '48px',
                        color: 'rgba(240,240,240,0.7)',
                    }}
                >
                    {description}
                </p>
                <div
                    style={{
                        display: 'flex',
                        flexDirection: 'row',
                        alignItems: 'center',
                        gap: '24px',
                        marginTop: 'auto',
                        color: primaryTextColor,
                    }}
                >
                    <span
                        style={{
                            fontSize: '46px',
                            fontWeight: 600,
                        }}
                    >
                        {siteName}
                    </span>
                </div>
            </div>
        </div>
    );
}