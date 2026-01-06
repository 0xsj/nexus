// app/[locale]/console/profile/page.tsx

export default function ProfileEditPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Profile</h1>
        <p className="text-sm text-muted-foreground">Manage your public profile</p>
      </div>

      {/* TODO: Profile edit form */}
      <div className="max-w-2xl space-y-4">
        <div className="h-24 w-24 bg-muted rounded-full" />
        <div className="h-10 bg-muted rounded" />
        <div className="h-10 bg-muted rounded" />
        <div className="h-24 bg-muted rounded" />
      </div>
    </div>
  );
}
