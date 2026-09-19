import { CustomCursor } from "@capa/ui/custom-cursor";
import type { Metadata } from "next";
import { Montserrat } from "next/font/google";
import { SnackbarProvider } from "@/components/snackbar";
import "./globals.css";

const montserrat = Montserrat({
  display: "swap",
  subsets: ["latin", "latin-ext"],
});

export const metadata: Metadata = {
  title: "CAPA Aday",
  description: "CAPA aday başvuru arayüzü",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="tr">
      <body className={montserrat.className}>
        <CustomCursor />
        <SnackbarProvider>{children}</SnackbarProvider>
      </body>
    </html>
  );
}
