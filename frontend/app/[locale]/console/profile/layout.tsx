// app/[locale]/(profile)/layout.tsx

interface ProfileLayoutProps {
  children: React.ReactNode;
}

export default function ProfileLayout({ children }: ProfileLayoutProps) {
  return (
    <div className="min-h-screen">
      <main>{children}</main>
    </div>
  );
}
