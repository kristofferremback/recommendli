import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'

export function OAuthCallback() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()

  useEffect(() => {
    const goto = searchParams.get('goto') || '/'
    navigate(goto, { replace: true })
  }, [searchParams, navigate])

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4" aria-busy="true" />
        <p className="text-gray-600">Completing authentication...</p>
      </div>
    </div>
  )
}
