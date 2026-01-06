// app/[locale]/console/settings/page.tsx

export default function SettingsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Settings</h1>
        <p className="text-sm text-muted-foreground">Manage your account settings</p>
      </div>

      {/* TODO: Settings sections */}
      <div className="max-w-2xl space-y-6">
        <div className="h-32 bg-muted rounded" />
        <div className="h-32 bg-muted rounded" />
        <div className="h-32 bg-muted rounded" />
      </div>
    </div>
  );
}
