// app/[locale]/(public)/roadmap/page.tsx

export default function RoadmapPage() {
  return (
    <div className="max-w-3xl mx-auto py-16 px-4">
      <h1 className="text-3xl font-bold mb-6">Roadmap</h1>
      <ul className="space-y-4 text-muted-foreground">
        <li>Phase 1: GitHub & LinkedIn integrations</li>
        <li>Phase 2: Education & certification providers</li>
        <li>Phase 3: Selective disclosure with ZK proofs</li>
        <li>Phase 4: Issuer portal for organizations</li>
        <li>Phase 5: Verification API</li>
        <li>Phase 6: Mobile wallet app</li>
      </ul>
    </div>
  );
}
