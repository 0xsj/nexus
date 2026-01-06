// app/[locale]/(public)/about/page.tsx

export default function AboutPage() {
  return (
    <div className="max-w-3xl mx-auto py-16 px-4">
      <h1 className="text-3xl font-bold mb-6">About Nexus</h1>
      <p className="text-muted-foreground">
        Nexus combines traditional platform integrations with decentralized identity standards,
        giving users portable, self-sovereign credentials they own and control.
      </p>
    </div>
  );
}
