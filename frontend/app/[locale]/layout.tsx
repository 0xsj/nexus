// app/[locale]/layout.tsx

import { QueryProvider } from "@/lib/query";

interface LocaleLayoutProps {
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}

export default async function LocaleLayout({ children, params }: LocaleLayoutProps) {
  const { locale } = await params;

  return (
    <QueryProvider>
      <div data-locale={locale}>{children}</div>
    </QueryProvider>
  );
}
