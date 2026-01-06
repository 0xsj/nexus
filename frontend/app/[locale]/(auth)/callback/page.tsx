// app/[locale]/(auth)/callback/page.tsx

export default function CallbackPage() {
  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-semibold">Completing sign in</h1>
        <p className="text-sm text-muted-foreground">
          Please wait while we complete your authentication...
        </p>
      </div>

      {/* TODO: OAuth callback handler */}
      <div className="flex justify-center">
        <div className="h-8 w-8 bg-muted rounded-full animate-pulse" />
      </div>
    </div>
  );
}
