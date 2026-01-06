// app/[locale]/console/page.tsx

export default function ConsolePage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Dashboard</h1>
        <p className="text-sm text-muted-foreground">Welcome back to your Nexus console</p>
      </div>

      {/* TODO: Dashboard widgets */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <div className="h-32 bg-muted rounded" />
        <div className="h-32 bg-muted rounded" />
        <div className="h-32 bg-muted rounded" />
      </div>
    </div>
  );
}
