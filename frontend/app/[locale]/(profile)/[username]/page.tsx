// app/[locale]/(profile)/[username]/page.tsx

interface ProfilePageProps {
  params: Promise<{ username: string }>;
}

export default async function ProfilePage({ params }: ProfilePageProps) {
  const { username } = await params;

  return (
    <div className="max-w-4xl mx-auto py-16 px-4">
      <div className="space-y-8">
        {/* Header */}
        <div className="flex items-center gap-6">
          <div className="h-24 w-24 bg-muted rounded-full" />
          <div className="space-y-2">
            <h1 className="text-2xl font-semibold">@{username}</h1>
            <p className="text-sm text-muted-foreground">Verified professional</p>
          </div>
        </div>

        {/* TODO: Verified credentials */}
        <div className="space-y-4">
          <h2 className="text-lg font-medium">Verified Credentials</h2>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="h-32 bg-muted rounded" />
            <div className="h-32 bg-muted rounded" />
          </div>
        </div>
      </div>
    </div>
  );
}
