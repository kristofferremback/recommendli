import { useCurrentUser } from '@/shared/api/queries'
import { DiscoveryPanel } from '@/features/discovery/components/DiscoveryPanel'
import { NowPlayingPanel } from '@/features/now-playing/components/NowPlayingPanel'
import { LibrarySummaryPanel } from '@/features/library-summary/components/LibrarySummaryPanel'

export function Dashboard() {
  const { data: user, isLoading } = useCurrentUser()

  if (isLoading) {
    return <LoadingDashboard />
  }

  if (!user) {
    return <LoginPrompt />
  }

  return (
    <div className="min-h-screen">
      <nav className="bg-white/80 backdrop-blur-xl border-b border-slate-200/60 sticky top-0 z-50 shadow-sm">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-black bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 bg-clip-text text-transparent">
              Recommendli
            </h1>
            <div className="flex items-center gap-2">
              <div className="h-2 w-2 rounded-full bg-green-500 animate-pulse"></div>
              <span className="text-sm font-medium text-slate-600">{user.display_name}</span>
            </div>
          </div>
        </div>
      </nav>

      <div className="container mx-auto px-6 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <DiscoveryPanel />
          <NowPlayingPanel />
          <LibrarySummaryPanel />
        </div>
      </div>
    </div>
  )
}

function LoadingDashboard() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <div className="inline-block animate-spin rounded-full h-16 w-16 border-4 border-slate-200 border-t-blue-600" aria-busy="true" />
        <p className="mt-4 text-slate-600 font-medium">Loading your music...</p>
      </div>
    </div>
  )
}

function LoginPrompt() {
  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="max-w-md w-full">
        <div className="text-center mb-8">
          <h1 className="text-5xl font-black mb-3 bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 bg-clip-text text-transparent">
            Recommendli
          </h1>
          <p className="text-slate-600 text-lg">
            Discover your next favorite track
          </p>
        </div>

        <div className="bg-white/80 backdrop-blur-sm rounded-2xl shadow-2xl border border-slate-200/60 p-8">
          <div className="mb-6">
            <h2 className="text-2xl font-bold text-slate-900 mb-2">Get Started</h2>
            <p className="text-slate-600">
              Connect your Spotify account to unlock personalized music recommendations based on your listening habits.
            </p>
          </div>

          <a
            href="/recommendations/v1/spotify/auth/ui-redirect?url=/"
            className="block w-full bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 text-white font-bold py-4 px-6 rounded-xl shadow-lg hover:shadow-xl hover:scale-[1.02] active:scale-[0.98] transition-all duration-200 text-center group relative overflow-hidden"
          >
            <span className="relative z-10 flex items-center justify-center gap-2">
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.66 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.479.659.301 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.601.18-1.2.72-1.381 4.26-1.26 11.28-1.02 15.721 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z"/>
              </svg>
              Connect with Spotify
            </span>
            <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-1000"></div>
          </a>

          <p className="mt-4 text-xs text-slate-500 text-center">
            We'll never post to your Spotify or share your data
          </p>
        </div>
      </div>
    </div>
  )
}
