import { motion } from 'framer-motion'
import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import axios from 'axios'
import Card from '../components/Card'
import Button from '../components/Button'
import Input from '../components/Input'

export default function Login() {
  const navigate = useNavigate()
  const [form, setForm] = useState({ email: '', password: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await axios.post('/auth/login', form)
      localStorage.setItem('token', res.data.token)
      localStorage.setItem('user', JSON.stringify(res.data))
      navigate('/dashboard')
    } catch (err) {
      setError(err.response?.data?.message || 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 24,
      backgroundImage: 'linear-gradient(rgba(37,99,235,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(37,99,235,0.04) 1px, transparent 1px)',
      backgroundSize: '60px 60px',
    }}>
      <motion.div
        initial={{ opacity: 0, y: 25 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <Card style={{ padding: 40, width: '100%', maxWidth: 420 }}>
          <Link to="/" style={{ textDecoration: 'none', display: 'block', textAlign: 'center' }}>
            <span style={{ fontFamily: 'Lobster, cursive', fontSize: 32, color: '#38bdf8' }}>Relay</span>
          </Link>

          <h2 style={{ fontSize: 24, fontWeight: 700, textAlign: 'center', marginTop: 24, marginBottom: 4 }}>Welcome back</h2>
          <p style={{ color: '#94a3b8', fontSize: 14, textAlign: 'center', marginBottom: 32 }}>Log in to your account</p>

          {error && (
            <motion.div
              initial={{ opacity: 0, x: -10 }}
              animate={{ opacity: 1, x: 0 }}
              style={{
                marginBottom: 20, padding: 12, borderRadius: 10,
                fontSize: 13, color: '#f87171',
                background: 'rgba(248, 113, 113, 0.08)',
                border: '1px solid rgba(248, 113, 113, 0.2)',
              }}
            >
              {error}
            </motion.div>
          )}

          <form onSubmit={handleSubmit}>
            <Input
              label="Email"
              type="email"
              placeholder="you@example.com"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              required
            />
            <Input
              label="Password"
              type="password"
              placeholder="••••••••"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              required
            />
            <Button type="submit" disabled={loading} style={{ width: '100%', marginTop: 8 }}>
              {loading ? 'Logging in...' : 'Log In'}
            </Button>
          </form>

          <p style={{ textAlign: 'center', fontSize: 13, color: '#94a3b8', marginTop: 24 }}>
            Don't have an account?{' '}
            <Link to="/signup" style={{ color: '#2563eb', fontWeight: 600, textDecoration: 'none' }}>Sign up</Link>
          </p>
        </Card>
      </motion.div>
    </div>
  )
}
