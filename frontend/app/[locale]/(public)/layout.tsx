// app/[locale]/(public)/layout.tsx

interface PublicLayoutProps {
  children: React.ReactNode;
}

export default function PublicLayout({ children }: PublicLayoutProps) {
  return (
    <div className="min-h-screen flex flex-col">
      {/* TODO: Navbar */}
      <header className="h-16 border-b">
        <nav className="h-full max-w-7xl mx-auto px-4 flex items-center justify-between">
          <span className="font-semibold">Nexus</span>
          <div className="flex items-center gap-4">{/* TODO: Nav links, auth buttons */}</div>
        </nav>
      </header>

      <main className="flex-1">{children}</main>

      {/* TODO: Footer */}
      <footer className="h-16 border-t">
        <div className="h-full max-w-7xl mx-auto px-4 flex items-center justify-center text-sm text-muted-foreground">
          © {new Date().getFullYear()} Nexus
        </div>
      </footer>
    </div>
  );
}
