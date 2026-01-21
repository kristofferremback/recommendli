import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Dashboard } from '@/routes/index'
import { AuthErrorBoundary } from '@/shared/components/AuthErrorBoundary'

function App() {
  return (
    <BrowserRouter>
      <AuthErrorBoundary>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthErrorBoundary>
    </BrowserRouter>
  )
}

export default App
