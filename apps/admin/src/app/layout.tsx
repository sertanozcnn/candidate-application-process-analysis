import { CustomCursor } from "@capa/ui/custom-cursor";
import type { Metadata } from "next";
import { Montserrat } from "next/font/google";
import "./globals.css";

const montserrat = Montserrat({
  display: "swap",
  subsets: ["latin", "latin-ext"],
});

export const metadata: Metadata = {
  title: "CAPA Admin",
  description: "CAPA yönetici arayüzü",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="tr">
      <body className={montserrat.className}>
        <CustomCursor />
        {children}
      </body>
    </html>
  );
}
