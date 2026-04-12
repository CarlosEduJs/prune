import { Geist, Geist_Mono, Inter, Space_Grotesk, DM_Sans, Fira_Code } from "next/font/google"

import "./globals.css"
import { ThemeProvider } from "@/components/theme-provider"
import { cn } from "@/lib/utils";

const firaCodeFiraCode = Fira_Code({subsets:['cyrillic','cyrillic-ext','greek','greek-ext','latin','latin-ext','symbols2'],weight:['300','400','500','600','700'],variable:'--font-fira-code'});

const dmSansDmSans = DM_Sans({subsets:['latin','latin-ext'],weight:['100','1000','200','300','400','500','600','700','800','900'],variable:'--font-dm-sans'});

const spaceGroteskSpaceGrotesk = Space_Grotesk({subsets:['latin','latin-ext','vietnamese'],weight:['300','400','500','600','700'],variable:'--font-space-grotesk'});

const inter = Inter({subsets:['latin'],variable:'--font-sans'})

const fontMono = Geist_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
})

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={cn("antialiased", fontMono.variable, inter.variable, spaceGroteskSpaceGrotesk.variable, dmSansDmSans.variable, firaCodeFiraCode.variable)}
    >
      <body>
        <ThemeProvider>{children}</ThemeProvider>
      </body>
    </html>
  )
}
