import { motion, AnimatePresence } from 'framer-motion'
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import axios from 'axios'
import Navbar from '../components/Navbar'
import Card from '../components/Card'
import Button from '../components/Button'
import Input from '../components/Input'

const typeColors = {
  'React': '#22d3ee', 'Vue': '#34d399', 'Angular': '#f87171', 'Svelte': '#fb923c',
  'Next.js': '#f8fafc', 'Node.js': '#4ade80', 'Go': '#60a5fa', 'Python': '#facc15',
  'Java (Maven)': '#fbbf24', 'Java (Gradle)': '#fbbf24', 'Static HTML': '#94a3b8',
}

export default function Dashboard() {
  const navigate = useNavigate()
  const [projects, setProjects] = useState([])
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [newProject, setNewProject] = useState({ projectName: '', repoUrl: '' })
  const [createError, setCreateError] = useState('')
  const [creating, setCreating] = useState(false)
  const token = localStorage.getItem('token')

  useEffect(() => {
    if (!token) { navigate('/login'); return }
    fetchProjects()
  }, [])

  const fetchProjects = async () => {
    try {
      const res = await axios.get('/api/projects', { headers: { Authorization: `Bearer ${token}` } })
      setProjects(res.data.projects || [])
    } catch { navigate('/login') }
    finally { setLoading(false) }
  }

  const handleCreate = async (e) => {
    e.preventDefault()
    setCreateError('')
    setCreating(true)
    try {
      await axios.post('/api/projects', newProject, { headers: { Authorization: `Bearer ${token}` } })
      setShowModal(false)
      setNewProject({ projectName: '', repoUrl: '' })
      fetchProjects()
    } catch (err) {
      setCreateError(err.response?.data?.message || 'Failed to create project')
    } finally { setCreating(false) }
  }

  return (
    <div style={{ minHeight: '100vh', backgroundImage: 'linear-gradient(rgba(37,99,235,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(37,99,235,0.04) 1px, transparent 1px)', backgroundSize: '60px 60px' }}>
      <Navbar showAuth={false} />

      <div style={{ maxWidth: 1200, margin: '0 auto', padding: '40px 24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 40 }}>
          <div>
            <h2 style={{ fontSize: 28, fontWeight: 800 }}>Your Projects</h2>
            <p style={{ color: '#94a3b8', marginTop: 4, fontSize: 14 }}>{projects.length} project{projects.length !== 1 ? 's' : ''}</p>
          </div>
          <Button onClick={() => setShowModal(true)}>+ New Project</Button>
        </div>

        {loading ? (
          <p style={{ textAlign: 'center', padding: 80, color: '#94a3b8' }}>Loading...</p>
        ) : projects.length === 0 ? (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
            <Card style={{ padding: 64, textAlign: 'center' }}>
              <div style={{ fontSize: 48, marginBottom: 16 }}>🚀</div>
              <h3 style={{ fontSize: 20, fontWeight: 700, marginBottom: 8 }}>No projects yet</h3>
              <p style={{ color: '#94a3b8', marginBottom: 24 }}>Create your first project to get started</p>
              <Button onClick={() => setShowModal(true)}>Create Project</Button>
            </Card>
          </motion.div>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(340px, 1fr))', gap: 20 }}>
            {projects.map((p, i) => (
              <motion.div key={p.id} initial={{ opacity: 0, y: 15 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.05 }}>
                <Card hover onClick={() => navigate(`/projects/${p.id}`)} style={{ padding: 24, cursor: 'pointer' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 16 }}>
                    <h3 style={{ fontSize: 17, fontWeight: 700, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: '70%' }}>{p.projectName}</h3>
                    <span style={{
                      fontSize: 11, fontWeight: 600, padding: '4px 10px', borderRadius: 8,
                      background: 'rgba(37,99,235,0.08)',
                      color: typeColors[p.projectType] || '#94a3b8',
                    }}>
                      {p.projectType}
                    </span>
                  </div>
                  <p style={{ fontSize: 12, color: '#64748b', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', marginBottom: 16 }}>{p.repoUrl}</p>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12, color: '#64748b' }}>
                    <span>{new Date(p.createdAt).toLocaleDateString()}</span>
                    <span style={{ color: '#2563eb' }}>→</span>
                  </div>
                </Card>
              </motion.div>
            ))}
          </div>
        )}
      </div>

      {/* Modal */}
      <AnimatePresence>
        {showModal && (
          <motion.div
            initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
            onClick={() => setShowModal(false)}
            style={{
              position: 'fixed', inset: 0, zIndex: 100,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              background: 'rgba(0,0,0,0.6)', backdropFilter: 'blur(4px)', padding: 24,
            }}
          >
            <motion.div
              initial={{ scale: 0.92, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} exit={{ scale: 0.92, opacity: 0 }}
              onClick={(e) => e.stopPropagation()}
            >
              <Card style={{ padding: 32, width: '100%', maxWidth: 420 }}>
                <h3 style={{ fontSize: 20, fontWeight: 700, marginBottom: 24 }}>New Project</h3>
                {createError && (
                  <div style={{ marginBottom: 16, padding: 12, borderRadius: 10, fontSize: 13, color: '#f87171', background: 'rgba(248,113,113,0.08)', border: '1px solid rgba(248,113,113,0.2)' }}>
                    {createError}
                  </div>
                )}
                <form onSubmit={handleCreate}>
                  <Input label="Project Name" placeholder="My Awesome App" value={newProject.projectName} onChange={(e) => setNewProject({ ...newProject, projectName: e.target.value })} required />
                  <Input label="GitHub URL" type="url" placeholder="https://github.com/user/repo" value={newProject.repoUrl} onChange={(e) => setNewProject({ ...newProject, repoUrl: e.target.value })} required />
                  <div style={{ display: 'flex', gap: 12, marginTop: 8 }}>
                    <Button variant="secondary" type="button" onClick={() => setShowModal(false)} style={{ flex: 1 }}>Cancel</Button>
                    <Button type="submit" disabled={creating} style={{ flex: 1 }}>{creating ? 'Creating...' : 'Create'}</Button>
                  </div>
                </form>
              </Card>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
