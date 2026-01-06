// app/[locale]/(public)/page.tsx

export default function LandingPage() {
  return (
    <div className="flex flex-col items-center justify-center py-24 px-4">
      <h1 className="text-4xl font-bold mb-4">Verified Identity & Credentials</h1>
      <p className="text-lg text-muted-foreground text-center max-w-2xl">
        A professional profile where everything is verified and cryptographically provable — not
        self-reported.
      </p>
    </div>
  );
}
