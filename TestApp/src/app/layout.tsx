import './globals.css'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'OmniBase TestApp',
  description: 'Test all authentication and identity features',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>
        <main style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
          {children}
        </main>
      </body>
    </html>
  )
}
