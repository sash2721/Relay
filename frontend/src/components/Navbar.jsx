import { useNavigate, Link } from 'react-router-dom'
import { motion } from 'framer-motion'

export default function Navbar({ showAuth = true, showBack = false, backTo = '/dashboard', backLabel = '← Back' }) {
  const navigate = useNavigate()
  const token = localStorage.getItem('token')
  const user = JSON.parse(localStorage.getItem('user') || '{}')

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    navigate('/')
  }

  return (
    <nav style={{
      position: 'sticky', top: 0, zIndex: 50,
      backdropFilter: 'blur(12px)',
      background: 'rgba(10, 22, 40, 0.85)',
      borderBottom: '1px solid rgba(37, 99, 235, 0.1)',
    }}>
      <div style={{
        maxWidth: 1200, margin: '0 auto',
        padding: '16px 24px',
        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
      }}>
        <Link to={token ? '/dashboard' : '/'} style={{ textDecoration: 'none' }}>
          <span style={{ fontFamily: 'Lobster, cursive', fontSize: 28, color: '#38bdf8' }}>Relay</span>
        </Link>

        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          {showBack && (
            <motion.button
              onClick={() => navigate(backTo)}
              whileTap={{ scale: 0.95 }}
              style={{
                background: 'none', border: 'none', color: '#94a3b8',
                fontSize: 14, cursor: 'pointer',
              }}
            >
              {backLabel}
            </motion.button>
          )}

          {showAuth && !token && (
            <>
              <motion.button
                onClick={() => navigate('/login')}
                whileHover={{ color: '#fff' }}
                style={{ background: 'none', border: 'none', color: '#94a3b8', fontSize: 14, cursor: 'pointer' }}
              >
                Log In
              </motion.button>
              <motion.button
                onClick={() => navigate('/signup')}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                style={{
                  background: 'linear-gradient(135deg, #2563eb, #1d4ed8)',
                  border: 'none', borderRadius: 10, padding: '8px 20px',
                  color: '#fff', fontSize: 14, fontWeight: 600, cursor: 'pointer',
                }}
              >
                Get Started
              </motion.button>
            </>
          )}

          {token && (
            <>
              <span style={{ fontSize: 13, color: '#94a3b8' }}>Hey, {user.name || 'there'}</span>
              <motion.button
                onClick={handleLogout}
                whileTap={{ scale: 0.95 }}
                style={{
                  background: 'none',
                  border: '1px solid rgba(37, 99, 235, 0.25)',
                  borderRadius: 8, padding: '6px 14px',
                  color: '#94a3b8', fontSize: 12, cursor: 'pointer',
                }}
              >
                Logout
              </motion.button>
            </>
          )}
        </div>
      </div>
    </nav>
  )
}
