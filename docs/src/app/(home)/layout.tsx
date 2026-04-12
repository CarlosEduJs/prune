export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <main className="flex h-screen w-screen flex-col gap-4 bg-linear-to-br from-primary/10 dark:from-primary/5 to-secondary/10 p-4">
      {children}
    </main>
  );
}
