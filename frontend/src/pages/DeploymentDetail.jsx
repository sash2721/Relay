import { motion } from 'framer-motion'
import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import axios from 'axios'
import Navbar from '../components/Navbar'
import Card from '../components/Card'

const statusSteps = ['pending', 'cloning', 'detecting', 'building', 'deploying', 'live']

const statusColors = {
  pending: '#facc15', cloning: '#60a5fa', detecting: '#a855f7',
  building: '#fb923c', deploying: '#22d3ee', live: '#4ade80', failed: '#f87171',
}

export default function DeploymentDetail() {
  const { projectId, deploymentId } = useParams()
  const navigate = useNavigate()
  const [deployment, setDeployment] = useState(null)
  const [logs, setLogs] = useState([])
  const [streaming, setStreaming] = useState(false)
  const logsEndRef = useRef(null)
  const token = localStorage.getItem('token')

  useEffect(() => {
    if (!token) { navigate('/login'); return }
    fetchDeployment()
  }, [])

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  useEffect(() => {
    if (deployment?.status === 'live' || deployment?.status === 'failed') return
    const interval = setInterval(fetchDeployment, 3000)
    return () => clearInterval(interval)
  }, [deployment?.status])

  const fetchDeployment = async () => {
    try {
      const res = await axios.get(`/api/projects/${projectId}/deployments/${deploymentId}`, { headers: { Authorization: `Bearer ${token}` } })
      setDeployment(res.data)
    } catch { navigate(`/projects/${projectId}`) }
  }

  const connectLogs = () => {
    const es = new EventSource(`/api/projects/${projectId}/deployments/${deploymentId}/logs?token=${token}`)
    es.onopen = () => setStreaming(true)
    es.onmessage = (e) => setLogs((prev) => [...prev, e.data])
    es.onerror = () => {
      setStreaming(false)
      es.close()
      // fetch full logs after stream ends
      setTimeout(fetchCompletedLogs, 1000)
    }

    return es
  }

  // cleanup EventSource on unmount
  useEffect(() => {
    const es = connectLogs()
    return () => es.close()
  }, [])

  const fetchCompletedLogs = async () => {
    try {
      const res = await axios.get(`/api/projects/${projectId}/deployments/${deploymentId}/logs`, { headers: { Authorization: `Bearer ${token}` } })
      if (Array.isArray(res.data)) setLogs(res.data)
    } catch {}
  }

  const currentStep = deployment ? statusSteps.indexOf(deployment.status) : 0
  const color = statusColors[deployment?.status] || '#94a3b8'

  return (
    <div style={{ minHeight: '100vh', backgroundImage: 'linear-gradient(rgba(37,99,235,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(37,99,235,0.04) 1px, transparent 1px)', backgroundSize: '60px 60px' }}>
      <Navbar showAuth={false} showBack backTo={`/projects/${projectId}`} backLabel="← Back to Project" />

      <div style={{ maxWidth: 960, margin: '0 auto', padding: '40px 24px' }}>
        {/* Status Card */}
        <motion.div initial={{ opacity: 0, y: 15 }} animate={{ opacity: 1, y: 0 }}>
          <Card style={{ padding: 32, marginBottom: 32 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 16, marginBottom: 24 }}>
              <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
                  <h2 style={{ fontSize: 24, fontWeight: 800 }}>Deployment</h2>
                  <span style={{
                    fontSize: 11, fontWeight: 700, padding: '4px 12px', borderRadius: 8,
                    background: `${color}15`, color: color, border: `1px solid ${color}30`,
                  }}>
                    {deployment?.status || 'loading'}
                  </span>
                </div>
                <p style={{ fontSize: 12, color: '#64748b', fontFamily: 'monospace' }}>{deploymentId}</p>
              </div>
              {deployment?.deployedUrl && (
                <a href={`http://${deployment.deployedUrl}:8080`} target="_blank" rel="noopener noreferrer"
                  style={{
                    background: 'linear-gradient(135deg, #2563eb, #1d4ed8)',
                    borderRadius: 10, padding: '10px 20px',
                    color: '#fff', fontSize: 13, fontWeight: 700, textDecoration: 'none',
                    display: 'inline-block',
                  }}>
                  🌐 Visit Site
                </a>
              )}
            </div>

            {/* Progress */}
            {deployment?.status !== 'failed' && (
              <div style={{ display: 'flex', alignItems: 'center', gap: 4, overflowX: 'auto', paddingBottom: 8 }}>
                {statusSteps.map((step, i) => (
                  <div key={step} style={{ display: 'flex', alignItems: 'center' }}>
                    <motion.div
                      animate={i === currentStep ? { scale: [1, 1.15, 1] } : {}}
                      transition={{ duration: 1.5, repeat: Infinity }}
                      style={{
                        width: 32, height: 32, borderRadius: '50%',
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        fontSize: 12, fontWeight: 700,
                        background: i <= currentStep ? '#2563eb' : 'transparent',
                        border: i <= currentStep ? '2px solid #2563eb' : '2px solid rgba(37,99,235,0.15)',
                        color: i <= currentStep ? '#fff' : '#64748b',
                      }}
                    >
                      {i <= currentStep ? '✓' : i + 1}
                    </motion.div>
                    {i < statusSteps.length - 1 && (
                      <div style={{ width: 24, height: 2, background: i < currentStep ? '#2563eb' : 'rgba(37,99,235,0.12)' }} />
                    )}
                  </div>
                ))}
              </div>
            )}

            {deployment?.failureReason && (
              <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} style={{
                marginTop: 16, padding: 16, borderRadius: 12,
                fontSize: 13, color: '#f87171', lineHeight: 1.6,
                background: 'rgba(248,113,113,0.05)', border: '1px solid rgba(248,113,113,0.15)',
              }}>
                <strong>Error: </strong>{deployment.failureReason}
              </motion.div>
            )}
          </Card>
        </motion.div>

        {/* Logs */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <h3 style={{ fontSize: 20, fontWeight: 700 }}>Build Logs</h3>
          {streaming && (
            <motion.div
              animate={{ opacity: [1, 0.4, 1] }}
              transition={{ duration: 1.5, repeat: Infinity }}
              style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 12, color: '#4ade80' }}
            >
              <div style={{ width: 7, height: 7, borderRadius: '50%', background: '#4ade80' }} />
              Live
            </motion.div>
          )}
        </div>

        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
          <Card style={{ padding: 24, maxHeight: 500, overflowY: 'auto', fontFamily: 'monospace', fontSize: 12, lineHeight: 1.8 }}>
            {logs.length === 0 ? (
              <div style={{ textAlign: 'center', padding: 40, color: '#64748b' }}>
                {streaming ? 'Waiting for logs...' : 'No logs available'}
              </div>
            ) : (
              logs.map((line, i) => (
                <div key={i} style={{ color: '#94a3b8', padding: '1px 0' }}>
                  <span style={{ color: '#1a3a6b', marginRight: 12, userSelect: 'none' }}>{String(i + 1).padStart(3, '0')}</span>
                  {line}
                </div>
              ))
            )}
            <div ref={logsEndRef} />
          </Card>
        </motion.div>
      </div>
    </div>
  )
}
