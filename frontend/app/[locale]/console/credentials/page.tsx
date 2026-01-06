// app/[locale]/console/credentials/page.tsx

export default function CredentialsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Credentials</h1>
        <p className="text-sm text-muted-foreground">Your verified credentials and certificates</p>
      </div>

      {/* TODO: Credential list */}
      <div className="grid gap-4 md:grid-cols-2">
        <div className="h-48 bg-muted rounded" />
        <div className="h-48 bg-muted rounded" />
        <div className="h-48 bg-muted rounded" />
      </div>
    </div>
  );
}
