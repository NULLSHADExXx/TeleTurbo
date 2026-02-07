import { useState, useEffect } from 'react'
import Login from './components/Login'
import Dashboard from './components/Dashboard'

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    checkAuthStatus()
  }, [])

  const checkAuthStatus = async () => {
    try {
      // Check if user is already authenticated
      if (window.go?.main?.App?.IsAuthenticated) {
        const auth = await window.go.main.App.IsAuthenticated()
        setIsAuthenticated(auth)
      }
    } catch (error) {
      console.log('Not authenticated yet')
    }
    setIsLoading(false)
  }

  const handleLoginSuccess = () => {
    setIsAuthenticated(true)
  }

  const handleLogout = () => {
    setIsAuthenticated(false)
  }

  if (isLoading) {
    return (
      <div className="loading-screen">
        <div className="loading-spinner"></div>
        <p>Loading TeleTurbo...</p>
      </div>
    )
  }

  return (
    <div className="app">
      {!isAuthenticated ? (
        <Login onLoginSuccess={handleLoginSuccess} />
      ) : (
        <Dashboard onLogout={handleLogout} />
      )}
    </div>
  )
}

export default App
