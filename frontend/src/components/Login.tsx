import { useState } from 'react'

interface LoginProps {
  onLoginSuccess: () => void
}

function Login({ onLoginSuccess }: LoginProps) {
  const [step, setStep] = useState<'credentials' | 'phone' | 'code' | 'password'>('credentials')
  const [appID, setAppID] = useState('')
  const [appHash, setAppHash] = useState('')
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleInitialize = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const result = await window.go.main.App.InitializeTelegramClient(
        parseInt(appID),
        appHash
      )
      
      if (result === 'CLIENT_INITIALIZED') {
        setStep('phone')
      } else {
        setError(result)
      }
    } catch (err) {
      setError('Failed to initialize client. Check your API credentials.')
    }
    
    setLoading(false)
  }

  const handleSendCode = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const result = await window.go.main.App.StartLogin(phone)
      
      if (result === 'CODE_SENT') {
        setStep('code')
      } else if (result === 'LOGIN_SUCCESS') {
        onLoginSuccess()
      } else {
        setError(result)
      }
    } catch (err) {
      setError('Failed to send code. Check your phone number.')
    }
    
    setLoading(false)
  }

  const handleSubmitCode = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const result = await window.go.main.App.SubmitCode(code)
      
      if (result === 'LOGIN_SUCCESS') {
        onLoginSuccess()
      } else if (result === 'PASSWORD_REQUIRED') {
        setStep('password')
      } else {
        setError(result)
      }
    } catch (err) {
      setError('Invalid code. Please try again.')
    }
    
    setLoading(false)
  }

  const handleSubmitPassword = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const result = await window.go.main.App.SubmitPassword(password)
      
      if (result === 'LOGIN_SUCCESS') {
        onLoginSuccess()
      } else {
        setError(result)
      }
    } catch (err) {
      setError('Invalid password. Please try again.')
    }
    
    setLoading(false)
  }

  return (
    <div className="login-container">
      <div className="login-box">
        <div className="logo-section">
          <div className="logo-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
              <polyline points="7 10 12 15 17 10"></polyline>
              <line x1="12" y1="15" x2="12" y2="3"></line>
            </svg>
          </div>
          <h1>TeleTurbo</h1>
          <p>High-Speed Telegram Downloader</p>
        </div>

        {step === 'credentials' && (
          <form onSubmit={handleInitialize} className="login-form">
            <div className="step-indicator">
              <div className="step-dot active"></div>
              <div className="step-dot"></div>
              <div className="step-dot"></div>
            </div>
            <h2>API Credentials</h2>
            <p className="form-subtitle">
              Enter your real API credentials from my.telegram.org
            </p>
            
            <div className="input-wrapper">
              <label>App ID</label>
              <div className="input-field">
                <input
                  type="number"
                  value={appID}
                  onChange={(e) => setAppID(e.target.value)}
                  placeholder="123456"
                  required
                />
              </div>
            </div>

            <div className="input-wrapper">
              <label>App Hash</label>
              <div className="input-field">
                <input
                  type="text"
                  value={appHash}
                  onChange={(e) => setAppHash(e.target.value)}
                  placeholder="a1b2c3d4e5f6..."
                  required
                />
              </div>
            </div>

            {error && <div className="error-message">{error}</div>}

            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? 'Initializing...' : 'Continue'}
            </button>
          </form>
        )}

        {step === 'phone' && (
          <form onSubmit={handleSendCode} className="login-form">
            <div className="step-indicator">
              <div className="step-dot active"></div>
              <div className="step-dot"></div>
              <div className="step-dot"></div>
            </div>
            <h2>Enter Phone Number</h2>
            <p className="form-subtitle">
              Format: +CountryCode Number (e.g., +14155551234)
            </p>
            
            <div className="input-wrapper">
              <label>Phone Number</label>
              <div className="input-field">
                <input
                  type="tel"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  placeholder="+1234567890"
                  required
                />
              </div>
            </div>

            {error && <div className="error-message">{error}</div>}

            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? 'Sending...' : 'Send Code'}
            </button>
          </form>
        )}

        {step === 'code' && (
          <form onSubmit={handleSubmitCode} className="login-form">
            <div className="step-indicator">
              <div className="step-dot active"></div>
              <div className="step-dot active"></div>
              <div className="step-dot"></div>
            </div>
            <h2>Verification Code</h2>
            <p className="form-subtitle">
              Enter the code we sent to your Telegram app
            </p>
            
            <div className="input-wrapper">
              <label>Code</label>
              <div className="input-field">
                <input
                  type="text"
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  placeholder="12345"
                  required
                />
              </div>
            </div>

            {error && <div className="error-message">{error}</div>}

            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? 'Verifying...' : 'Verify'}
            </button>
            
            <button 
              type="button" 
              className="btn-secondary"
              onClick={() => setStep('phone')}
            >
              Back
            </button>
          </form>
        )}

        {step === 'password' && (
          <form onSubmit={handleSubmitPassword} className="login-form">
            <div className="step-indicator">
              <div className="step-dot active"></div>
              <div className="step-dot active"></div>
              <div className="step-dot active"></div>
            </div>
            <h2>Cloud Password</h2>
            <p className="form-subtitle">
              Your account has a cloud password. Enter it to continue.
            </p>
            
            <div className="input-wrapper">
              <label>Password</label>
              <div className="input-field">
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Your cloud password"
                  required
                />
              </div>
            </div>

            {error && <div className="error-message">{error}</div>}

            <button type="submit" className="btn-primary" disabled={loading}>
              {loading ? 'Verifying...' : 'Login'}
            </button>
          </form>
        )}
      </div>
    </div>
  )
}

export default Login
