import React, { Component, ReactNode } from 'react'
import { toast } from 'sonner'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
}

/**
 * Error boundary that catches auth errors (401/403) and redirects to OAuth
 * Handles TanStack Query errors globally
 */
export class AuthErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: any): State {
    if (error?.status === 401 || error?.status === 403) {
      const currentUrl = window.location.href
      window.location.replace(
        `/recommendations/v1/spotify/auth/ui-redirect?url=${encodeURIComponent(currentUrl)}`
      )
    }
    return { hasError: true }
  }

  componentDidCatch(error: any, errorInfo: React.ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo)

    if (error?.status !== 401 && error?.status !== 403) {
      toast.error('Something went wrong', {
        description: error.message || 'Please try again'
      })
    }
  }

  render() {
    if (this.state.hasError) {
      return null
    }

    return this.props.children
  }
}
