// app/[locale]/(console)/layout.tsx

interface ConsoleLayoutProps {
  children: React.ReactNode;
}

export default function ConsoleLayout({ children }: ConsoleLayoutProps) {
  return (
    <div className="min-h-screen flex">
      {/* Sidebar */}
      <aside className="w-64 border-r hidden md:flex flex-col">
        <div className="h-16 border-b flex items-center px-6">
          <span className="font-semibold">Nexus</span>
        </div>
        <nav className="flex-1 p-4 space-y-2">
          {/* TODO: Nav items */}
          <div className="text-sm text-muted-foreground">Dashboard</div>
          <div className="text-sm text-muted-foreground">Credentials</div>
          <div className="text-sm text-muted-foreground">Integrations</div>
          <div className="text-sm text-muted-foreground">Profile</div>
          <div className="text-sm text-muted-foreground">Settings</div>
        </nav>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col">
        <header className="h-16 border-b flex items-center justify-between px-6">
          {/* TODO: Mobile menu, search, user menu */}
          <div className="md:hidden font-semibold">Nexus</div>
          <div className="flex items-center gap-4">{/* TODO: User avatar/menu */}</div>
        </header>
        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  );
}
