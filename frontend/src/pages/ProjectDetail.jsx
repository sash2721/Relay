import { motion } from 'framer-motion'
import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import axios from 'axios'
import Navbar from '../components/Navbar'
import Card from '../components/Card'
import Button from '../components/Button'

const statusStyles = {
  pending: { bg: 'rgba(250,204,21,0.1)', color: '#facc15', border: 'rgba(250,204,21,0.2)' },
  cloning: { bg: 'rgba(96,165,250,0.1)', color: '#60a5fa', border: 'rgba(96,165,250,0.2)' },
  detecting: { bg: 'rgba(168,85,247,0.1)', color: '#a855f7', border: 'rgba(168,85,247,0.2)' },
  building: { bg: 'rgba(251,146,60,0.1)', color: '#fb923c', border: 'rgba(251,146,60,0.2)' },
  deploying: { bg: 'rgba(34,211,238,0.1)', color: '#22d3ee', border: 'rgba(34,211,238,0.2)' },
  live: { bg: 'rgba(74,222,128,0.1)', color: '#4ade80', border: 'rgba(74,222,128,0.2)' },
  failed: { bg: 'rgba(248,113,113,0.1)', color: '#f87171', border: 'rgba(248,113,113,0.2)' },
}

export default function ProjectDetail() {
  const { projectId } = useParams()
  const navigate = useNavigate()
  const [project, setProject] = useState(null)
  const [deployments, setDeployments] = useState([])
  const [loading, setLoading] = useState(true)
  const [deploying, setDeploying] = useState(false)
  const [deployError, setDeployError] = useState('')
  const token = localStorage.getItem('token')

  useEffect(() => {
    if (!token) { navigate('/login'); return }
    fetchData()
  }, [])

  const fetchData = async () => {
    try {
      const [projRes, depRes] = await Promise.all([
        axios.get(`/api/projects/${projectId}`, { headers: { Authorization: `Bearer ${token}` } }),
        axios.get(`/api/projects/${projectId}/deployments`, { headers: { Authorization: `Bearer ${token}` } }),
      ])
      setProject(projRes.data)
      setDeployments(depRes.data.deployments || [])
    } catch { navigate('/dashboard') }
    finally { setLoading(false) }
  }

  const triggerDeploy = async () => {
    setDeployError('')
    setDeploying(true)
    try {
      await axios.post(`/api/projects/${projectId}/deployments`, {}, { headers: { Authorization: `Bearer ${token}` } })
      fetchData()
    } catch (err) {
      setDeployError(err.response?.data?.message || 'Deploy failed')
    } finally { setDeploying(false) }
  }

  if (loading) return <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#94a3b8' }}>Loading...</div>

  return (
    <div style={{ minHeight: '100vh', backgroundImage: 'linear-gradient(rgba(37,99,235,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(37,99,235,0.04) 1px, transparent 1px)', backgroundSize: '60px 60px' }}>
      <Navbar showAuth={false} showBack backTo="/dashboard" backLabel="← Back to Projects" />

      <div style={{ maxWidth: 1200, margin: '0 auto', padding: '40px 24px' }}>
        <motion.div initial={{ opacity: 0, y: 15 }} animate={{ opacity: 1, y: 0 }}>
          <Card style={{ padding: 32, marginBottom: 32 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 16 }}>
              <div>
                <h2 style={{ fontSize: 28, fontWeight: 800, marginBottom: 8 }}>{project?.projectName}</h2>
                <p style={{ fontSize: 13, color: '#64748b', marginBottom: 4 }}>{project?.repoUrl}</p>
                <span style={{ fontSize: 12, fontWeight: 600, color: '#14b8a6' }}>{project?.projectType}</span>
              </div>
              <Button onClick={triggerDeploy} disabled={deploying}>
                {deploying ? 'Deploying...' : '🚀 Deploy Now'}
              </Button>
            </div>
            {deployError && (
              <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} style={{ marginTop: 16, padding: 12, borderRadius: 10, fontSize: 13, color: '#f87171', background: 'rgba(248,113,113,0.08)', border: '1px solid rgba(248,113,113,0.2)' }}>
                {deployError}
              </motion.div>
            )}
          </Card>
        </motion.div>

        <h3 style={{ fontSize: 20, fontWeight: 700, marginBottom: 20 }}>Deployments</h3>

        {deployments.length === 0 ? (
          <Card style={{ padding: 48, textAlign: 'center' }}>
            <div style={{ fontSize: 40, marginBottom: 12 }}>📦</div>
            <p style={{ color: '#94a3b8' }}>No deployments yet. Hit Deploy to get started.</p>
          </Card>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {deployments.map((dep, i) => {
              const s = statusStyles[dep.status] || statusStyles.pending
              return (
                <motion.div key={dep.id} initial={{ opacity: 0, x: -15 }} animate={{ opacity: 1, x: 0 }} transition={{ delay: i * 0.05 }}>
                  <Card hover onClick={() => navigate(`/projects/${projectId}/deployments/${dep.id}`)} style={{ padding: 20, cursor: 'pointer' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 12 }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                        <span style={{
                          fontSize: 11, fontWeight: 700, padding: '4px 12px', borderRadius: 8,
                          background: s.bg, color: s.color, border: `1px solid ${s.border}`,
                        }}>
                          {dep.status}
                        </span>
                        <span style={{ fontSize: 13, color: '#64748b', fontFamily: 'monospace' }}>{dep.id.slice(0, 8)}</span>
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                        {dep.deployedUrl && (
                          <a href={`http://${dep.deployedUrl}:8080`} target="_blank" rel="noopener noreferrer"
                            onClick={(e) => e.stopPropagation()}
                            style={{ fontSize: 12, color: '#14b8a6', textDecoration: 'none' }}>
                            {dep.deployedUrl}
                          </a>
                        )}
                        <span style={{ fontSize: 12, color: '#64748b' }}>{new Date(dep.createdAt).toLocaleString()}</span>
                        <span style={{ color: '#2563eb' }}>→</span>
                      </div>
                    </div>
                    {dep.failureReason && (
                      <p style={{ marginTop: 12, fontSize: 12, color: '#f87171', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{dep.failureReason}</p>
                    )}
                  </Card>
                </motion.div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
