export const metadata = {
  title: 'EventPilot',
  description: 'Calendar and marketing agent',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
